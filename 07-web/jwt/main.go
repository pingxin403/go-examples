// main.go - JWT 示例
//
// 本示例演示 Go 中 JWT 的完整使用流程：
//   - 生成 HS256 Token
//   - 解析和验证 Token
//   - 自定义 Claims (UserID, Role)
//   - Token 过期
//   - HTTP 中间件实现 JWT 认证
//
// 安装依赖:
//   go get github.com/golang-jwt/jwt/v5

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// =========================================================
// 1. 配置
// =========================================================

// JWTSecret JWT 签名密钥（生产环境应从配置文件或环境变量读取）
var JWTSecret = []byte("my-very-secret-key-change-in-production")

// TokenExpiration Token 过期时间
const TokenExpiration = 30 * time.Minute

// =========================================================
// 2. 自定义 Claims
// =========================================================

// CustomClaims 自定义 JWT Claims，除标准字段外携带用户信息
type CustomClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// =========================================================
// 3. Token 生成
// =========================================================

// GenerateToken 生成 HS256 JWT Token
func GenerateToken(userID uint, role string) (string, error) {
	// 创建自定义 Claims
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			// 签发者
			Issuer: "go-examples",
			// 主题
			Subject: fmt.Sprintf("user_%d", userID),
			// 受众
			Audience: jwt.ClaimStrings{"api-server"},
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
			// 生效时间
			NotBefore: jwt.NewNumericDate(time.Now()),
			// 签发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// Token ID（用于防重放攻击）
			ID: fmt.Sprintf("token_%d", time.Now().UnixNano()),
		},
	}

	// 创建 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", fmt.Errorf("签名 Token 失败: %w", err)
	}

	return tokenString, nil
}

// =========================================================
// 4. Token 解析与验证
// =========================================================

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析 Token 失败: %w", err)
	}

	// 提取 Claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("无效的 Token")
	}

	return claims, nil
}

// =========================================================
// 5. JWT 中间件（HTTP）
// =========================================================

// AuthMiddleware JWT 认证 HTTP 中间件
// 从 Authorization 请求头中提取并验证 Token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取 Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"缺少 Authorization 请求头"}`, http.StatusUnauthorized)
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Authorization 格式应为 Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		// 解析和验证 Token
		claims, err := ParseToken(parts[1])
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// 将 Claims 信息注入请求上下文（使用自定义 header 模拟——生产中应使用 context.WithValue）
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		r.Header.Set("X-User-Role", claims.Role)

		// 调用下一个 handler
		next(w, r)
	}
}

// =========================================================
// 6. 基于角色的授权检查
// =========================================================

// RequireRole 基于角色的访问控制中间件
func RequireRole(role string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Header.Get("X-User-Role")
			if userRole != role {
				http.Error(w, fmt.Sprintf(`{"error":"需要 %s 角色，当前角色为 %s"}`, role, userRole), http.StatusForbidden)
				return
			}
			next(w, r)
		})
	}
}

// =========================================================
// 7. HTTP 处理器
// =========================================================

// loginHandler 模拟登录接口，返回 JWT Token
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"请使用 POST 方法"}`, http.StatusMethodNotAllowed)
		return
	}

	// 模拟：从请求中获取用户名和密码
	username := r.FormValue("username")
	password := r.FormValue("password")

	// 模拟用户验证（生产环境应从数据库验证）
	var userID uint
	var role string

	switch {
	case username == "admin" && password == "admin123":
		userID = 1
		role = "admin"
	case username == "user" && password == "user123":
		userID = 2
		role = "user"
	default:
		http.Error(w, `{"error":"用户名或密码错误"}`, http.StatusUnauthorized)
		return
	}

	// 生成 Token
	token, err := GenerateToken(userID, role)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"生成 Token 失败: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"token":"%s","expires_in":"%s"}`, token, TokenExpiration.String())
}

// profileHandler 获取用户信息（需要认证）
func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"user_id":%s,"role":"%s","message":"这是受保护的个人信息"}`, userID, role)
}

// adminHandler 管理接口（仅 admin 角色可访问）
func adminHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"user_id":%s,"message":"欢迎管理员！这是管理后台"}`, userID)
}

// publicHandler 公开接口，无需认证
func publicHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message":"这是公开接口，无需认证"}`)
}

// =========================================================
// 8. Token 验证和调试函数
// =========================================================

func demoTokenLifecycle() {
	log.Println("========== JWT Token 生命周期演示 ==========")

	// 生成 Token
	tokenStr, err := GenerateToken(1001, "admin")
	if err != nil {
		log.Fatalf("生成 Token 失败: %v", err)
	}
	log.Printf("生成的 Token: %s", tokenStr)
	log.Printf("Token 长度: %d 字符", len(tokenStr))

	// 解析 Token
	claims, err := ParseToken(tokenStr)
	if err != nil {
		log.Fatalf("解析 Token 失败: %v", err)
	}

	log.Printf("解析 Token 成功:")
	log.Printf("  UserID: %d", claims.UserID)
	log.Printf("  Role:   %s", claims.Role)
	log.Printf("  签发者: %s", claims.Issuer)
	log.Printf("  主题:   %s", claims.Subject)
	log.Printf("  过期时间: %s", claims.ExpiresAt.Format(time.RFC3339))

	// 验证 Token 是否过期
	timeLeft := time.Until(claims.ExpiresAt.Time)
	log.Printf("  剩余有效期: %v", timeLeft)

	// 演示不同角色的 Token
	roles := []string{"admin", "user", "guest"}
	for _, role := range roles {
		tok, _ := GenerateToken(uint(1000+len(role)), role)
		log.Printf("[%s] Token (前20字符): %s...", role, tok[:20])
	}

	// 演示解析带过期时间的 Token
	fmt.Println()
	log.Println("=== Token 过期验证 ===")
	expiredClaims := CustomClaims{
		UserID: 999,
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1 小时前过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredStr, _ := expiredToken.SignedString(JWTSecret)

	_, err = ParseToken(expiredStr)
	if err != nil {
		log.Printf("过期的 Token 验证结果: %v (符合预期)", err)
	}
}

// =========================================================
// 9. 主函数
// =========================================================

func main() {
	// =========================================================
	// 控制台演示 Token 生命周期
	// =========================================================
	demoTokenLifecycle()

	fmt.Println()

	// =========================================================
	// 启动 HTTP 服务器展示 JWT 中间件
	// =========================================================
	log.Println("========== JWT HTTP 服务器启动 ==========")

	// 注册路由
	mux := http.NewServeMux()

	// 公开接口
	mux.HandleFunc("/api/public", publicHandler)

	// 登录接口（获取 Token）
	mux.HandleFunc("/api/login", loginHandler)

	// 受保护接口（需要有效 Token）
	mux.HandleFunc("/api/profile", AuthMiddleware(profileHandler))

	// 管理员接口（需要 admin 角色）
	mux.HandleFunc("/api/admin", RequireRole("admin")(adminHandler))

	// 启动服务器
	port := ":8080"
	log.Printf("服务器启动于 http://localhost%s", port)
	log.Println()
	log.Println("可用接口:")
	log.Println("  GET  /api/public      — 公开接口")
	log.Println("  POST /api/login       — 登录获取 Token")
	log.Println("       username=admin&password=admin123")
	log.Println("       username=user&password=user123")
	log.Println("  GET  /api/profile     — 获取个人信息（需 Token）")
	log.Println("       Header: Authorization: Bearer <token>")
	log.Println("  GET  /api/admin       — 管理接口（需 admin 角色）")
	log.Println()
	log.Println("快速测试（命令行）:")
	log.Println("  # 获取 Token:")
	log.Println(`  curl -X POST http://localhost:8080/api/login -d "username=admin&password=admin123"`)
	log.Println()
	log.Println("  # 访问受保护接口:")
	log.Println(`  TOKEN=$(curl -s -X POST http://localhost:8080/api/login -d "username=admin&password=admin123" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)`)
	log.Println(`  curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/profile`)
	log.Println()

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}