// main_test.go — JWT Token 单元测试
//
// 测试内容：
//   - GenerateToken 生成 HS256 JWT
//   - ParseToken 解析和验证 Token
//   - 自定义 Claims（UserID, Role）
//   - Token 过期验证
//   - HTTP 中间件（AuthMiddleware, RequireRole）
//   - HTTP 处理器（loginHandler, profileHandler, adminHandler, publicHandler）

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ---------- Token 生成与解析测试 ----------

// TestGenerateToken 测试 Token 生成功能
func TestGenerateToken(t *testing.T) {
	t.Run("生成管理员 Token", func(t *testing.T) {
		token, err := GenerateToken(1, "admin")
		if err != nil {
			t.Fatalf("GenerateToken 失败: %v", err)
		}
		if token == "" {
			t.Fatal("Token 不应为空")
		}
		// Token 应由点号分隔的三部分组成
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Errorf("JWT 应有 3 段，实际得到 %d 段", len(parts))
		}
	})

	t.Run("生成普通用户 Token", func(t *testing.T) {
		token, err := GenerateToken(2, "user")
		if err != nil {
			t.Fatalf("GenerateToken 失败: %v", err)
		}
		if token == "" {
			t.Fatal("Token 不应为空")
		}
	})
}

// TestParseToken 测试 Token 解析功能
func TestParseToken(t *testing.T) {
	t.Run("解析有效 Token", func(t *testing.T) {
		tokenStr, err := GenerateToken(1001, "admin")
		if err != nil {
			t.Fatalf("生成 Token 失败: %v", err)
		}

		claims, err := ParseToken(tokenStr)
		if err != nil {
			t.Fatalf("ParseToken 失败: %v", err)
		}

		if claims.UserID != 1001 {
			t.Errorf("预期 UserID=1001，实际得到 %d", claims.UserID)
		}
		if claims.Role != "admin" {
			t.Errorf("预期 Role=admin，实际得到 %s", claims.Role)
		}
		if claims.Issuer != "go-examples" {
			t.Errorf("预期 Issuer=go-examples，实际得到 %s", claims.Issuer)
		}
	})

	t.Run("解析过期的 Token", func(t *testing.T) {
		// 创建已过期 1 小时的 Token
		claims := CustomClaims{
			UserID: 999,
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredStr, err := expiredToken.SignedString(JWTSecret)
		if err != nil {
			t.Fatalf("签名过期 Token 失败: %v", err)
		}

		_, err = ParseToken(expiredStr)
		if err == nil {
			t.Error("过期 Token 应返回错误")
		}
	})

	t.Run("解析无效的 Token 字符串", func(t *testing.T) {
		_, err := ParseToken("invalid-token-string")
		if err == nil {
			t.Error("无效 Token 应返回错误")
		}
	})

	t.Run("解析空字符串", func(t *testing.T) {
		_, err := ParseToken("")
		if err == nil {
			t.Error("空 Token 应返回错误")
		}
	})

	t.Run("使用错误的密钥签名", func(t *testing.T) {
		claims := CustomClaims{
			UserID: 1,
			Role:   "admin",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		wrongSecret := []byte("wrong-secret")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString(wrongSecret)

		_, err := ParseToken(tokenStr)
		if err == nil {
			t.Error("使用错误密钥签名的 Token 应返回错误")
		}
	})

	t.Run("完整的 Token 生命周期", func(t *testing.T) {
		// 生成 → 解析 → 验证 Claims
		originalUserID := uint(42)
		originalRole := "editor"

		tokenStr, err := GenerateToken(originalUserID, originalRole)
		if err != nil {
			t.Fatalf("生成 Token 失败: %v", err)
		}

		claims, err := ParseToken(tokenStr)
		if err != nil {
			t.Fatalf("解析 Token 失败: %v", err)
		}

		if claims.UserID != originalUserID {
			t.Errorf("UserID 不匹配: 预期 %d, 实际 %d", originalUserID, claims.UserID)
		}
		if claims.Role != originalRole {
			t.Errorf("Role 不匹配: 预期 %s, 实际 %s", originalRole, claims.Role)
		}
		if !claims.ExpiresAt.Time.After(time.Now()) {
			t.Error("Token 应在有效期内")
		}
	})
}

// ---------- HTTP 中间件测试 ----------

// setupTestMux 创建测试用的 HTTP ServeMux
func setupTestMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/public", publicHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/profile", AuthMiddleware(profileHandler))
	mux.HandleFunc("/api/admin", RequireRole("admin")(adminHandler))
	return mux
}

// performRequest 执行 HTTP 请求并返回响应
func performRequest(handler http.Handler, method, path string, formData map[string]string) *http.Response {
	var req *http.Request
	if len(formData) > 0 {
		form := url.Values{}
		for k, v := range formData {
			form.Set(k, v)
		}
		req = httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w.Result()
}

// TestPublicHandler 测试公开接口
func TestPublicHandler(t *testing.T) {
	handler := setupTestMux()
	resp := performRequest(handler, "GET", "/api/public", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("预期 200，实际得到 %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["message"] != "这是公开接口，无需认证" {
		t.Errorf("返回信息不正确: %v", body["message"])
	}
}

// TestLoginHandler 测试登录接口
func TestLoginHandler(t *testing.T) {
	handler := setupTestMux()

	t.Run("管理员登录成功", func(t *testing.T) {
		resp := performRequest(handler, "POST", "/api/login", map[string]string{
			"username": "admin",
			"password": "admin123",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", resp.StatusCode)
		}

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		if body["token"] == "" {
			t.Error("登录成功后应返回 Token")
		}
	})

	t.Run("普通用户登录成功", func(t *testing.T) {
		resp := performRequest(handler, "POST", "/api/login", map[string]string{
			"username": "user",
			"password": "user123",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", resp.StatusCode)
		}

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		if body["token"] == "" {
			t.Error("登录成功后应返回 Token")
		}
	})

	t.Run("登录失败 — 错误密码", func(t *testing.T) {
		resp := performRequest(handler, "POST", "/api/login", map[string]string{
			"username": "admin",
			"password": "wrong",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", resp.StatusCode)
		}
	})

	t.Run("登录失败 — 不存在的用户", func(t *testing.T) {
		resp := performRequest(handler, "POST", "/api/login", map[string]string{
			"username": "nobody",
			"password": "anything",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", resp.StatusCode)
		}
	})

	t.Run("使用 GET 方法 — 应返回 405", func(t *testing.T) {
		resp := performRequest(handler, "GET", "/api/login", nil)
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("预期 405，实际得到 %d", resp.StatusCode)
		}
	})
}

// TestProfileHandler 测试受保护的个人信息接口
func TestProfileHandler(t *testing.T) {
	handler := setupTestMux()

	t.Run("无 Token — 应返回 401", func(t *testing.T) {
		resp := performRequest(handler, "GET", "/api/profile", nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", resp.StatusCode)
		}
	})

	t.Run("带有效 Token — 应返回 200", func(t *testing.T) {
		token, _ := GenerateToken(1, "admin")
		req := httptest.NewRequest("GET", "/api/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d: %s", w.Code, w.Body.String())
		}

		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)

		// user_id 在 JSON 中是数字类型
		userID, ok := body["user_id"].(float64)
		if !ok || userID != 1 {
			t.Errorf("预期 user_id=1，实际得到 %v (类型: %T)", body["user_id"], body["user_id"])
		}
		if body["role"] != "admin" {
			t.Errorf("预期 role=admin，实际得到 %v", body["role"])
		}
	})

	t.Run("Token 格式错误", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/profile", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", w.Code)
		}
	})
}

// TestAdminHandler 测试管理员专用接口
func TestAdminHandler(t *testing.T) {
	handler := setupTestMux()

	t.Run("admin 角色 — 应通过", func(t *testing.T) {
		token, _ := GenerateToken(1, "admin")
		req := httptest.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("user 角色 — 应返回 403", func(t *testing.T) {
		token, _ := GenerateToken(2, "user")
		req := httptest.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("预期 403，实际得到 %d", w.Code)
		}
	})

	t.Run("无 Token — 应返回 401", func(t *testing.T) {
		resp := performRequest(handler, "GET", "/api/admin", nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", resp.StatusCode)
		}
	})
}

// TestAuthMiddlewareErrorCases 测试中间件的边界情况
func TestAuthMiddlewareErrorCases(t *testing.T) {
	t.Run("空 Authorization 请求头", func(t *testing.T) {
		handler := AuthMiddleware(profileHandler)
		req := httptest.NewRequest("GET", "/profile", nil)
		req.Header.Set("Authorization", "")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", w.Code)
		}
	})

	t.Run("Bearer 前缀后无 Token", func(t *testing.T) {
		handler := AuthMiddleware(profileHandler)
		req := httptest.NewRequest("GET", "/profile", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", w.Code)
		}
	})

	t.Run("非 Bearer 格式", func(t *testing.T) {
		handler := AuthMiddleware(profileHandler)
		req := httptest.NewRequest("GET", "/profile", nil)
		req.Header.Set("Authorization", "Token abc123")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", w.Code)
		}
	})
}