// main_test.go — Gin Web 框架单元测试
//
// 测试内容：
//   - 路由处理器（listUsers, getUserByID, createUser, updateUser, deleteUser）
//   - 搜索功能（searchUsers）
//   - 认证中间件（AuthMiddleware）
//   - 公开接口（greetUser, listArticles, getArticleByID）
//   - 健康检查（/ping）

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestRouter 创建测试用的 Gin 引擎，设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 重新初始化测试数据
	initTestData()

	// API v1 分组 - 公开文章接口
	v1 := r.Group("/api/v1")
	{
		articles := v1.Group("/articles")
		{
			articles.GET("", listArticles)
			articles.GET("/:id", getArticleByID)
		}

		// 用户接口（需要认证）
		users := v1.Group("/users")
		users.Use(AuthMiddleware())
		{
			users.GET("", listUsers)
			users.GET("/:id", getUserByID)
			users.GET("/search", searchUsers)
			users.POST("", createUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}
	}

	// 公开路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	r.GET("/hello", greetUser)

	return r
}

// initTestData 重置测试数据到已知状态
func initTestData() {
	users = map[uint]User{
		1: {ID: 1, Username: "alice", Email: "alice@example.com", Age: 28},
		2: {ID: 2, Username: "bob", Email: "bob@example.com", Age: 35},
	}
	articles = map[uint]Article{}
	nextID = 3
}

// authHeader 生成测试用的 Authorization 请求头键值对
func authHeader() []string {
	return []string{"Authorization", "Bearer my-secret-token"}
}



// performRequest 执行 HTTP 请求并返回响应
func performRequest(r http.Handler, method, path string, body string, headers ...string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	// 设置额外的请求头（成对传入）
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------- 测试用例 ----------

// TestPing 测试健康检查接口
func TestPing(t *testing.T) {
	r := setupTestRouter()
	w := performRequest(r, "GET", "/ping", "")

	if w.Code != http.StatusOK {
		t.Errorf("预期状态码 200，实际得到 %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "pong" {
		t.Errorf("预期 message=pong，实际得到 %s", resp["message"])
	}
}

// TestGreetUser 测试问候接口
func TestGreetUser(t *testing.T) {
	r := setupTestRouter()

	tests := []struct {
		name   string
		path   string
		want   int
		prefix string
	}{
		{"带参数", "/hello?name=张三", http.StatusOK, "你好，张三"},
		{"默认参数", "/hello", http.StatusOK, "你好，游客"},
		{"空参数", "/hello?name=", http.StatusOK, "你好，"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(r, "GET", tt.path, "")
			if w.Code != tt.want {
				t.Errorf("预期状态码 %d，实际得到 %d", tt.want, w.Code)
			}
			body := w.Body.String()
			if !strings.HasPrefix(body, tt.prefix) {
				t.Errorf("预期以 %q 开头，实际得到 %q", tt.prefix, body)
			}
		})
	}
}

// TestListUsers 测试用户列表接口（需要认证）
func TestListUsers(t *testing.T) {
	r := setupTestRouter()

	t.Run("未认证 — 应返回 401", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users", "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", w.Code)
		}
	})

	t.Run("无效 Token — 应返回 403", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users", "", "Authorization", "Bearer bad-token")
		if w.Code != http.StatusForbidden {
			t.Errorf("预期 403，实际得到 %d", w.Code)
		}
	})

	t.Run("已认证 — 获取用户列表", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users", "", authHeader()...)

		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		count := resp["count"].(float64)
		if count != 2 {
			t.Errorf("预期 2 个用户，实际得到 %v", count)
		}
	})
}

// TestGetUserByID 测试获取单个用户
func TestGetUserByID(t *testing.T) {
	r := setupTestRouter()

	t.Run("获取存在的用户", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users/1", "", authHeader()...)
		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", w.Code)
		}

		var user User
		json.Unmarshal(w.Body.Bytes(), &user)
		if user.Username != "alice" {
			t.Errorf("预期用户名 alice，实际得到 %s", user.Username)
		}
	})

	t.Run("获取不存在的用户", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users/999", "", authHeader()...)
		if w.Code != http.StatusNotFound {
			t.Errorf("预期 404，实际得到 %d", w.Code)
		}
	})

	t.Run("无效 ID 格式", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users/abc", "", authHeader()...)
		if w.Code != http.StatusBadRequest {
			t.Errorf("预期 400，实际得到 %d", w.Code)
		}
	})
}

// TestCreateUser 测试创建用户
func TestCreateUser(t *testing.T) {
	r := setupTestRouter()

	t.Run("创建有效用户", func(t *testing.T) {
		body := `{"username":"charlie","email":"charlie@example.com","age":25}`
		w := performRequest(r, "POST", "/api/v1/users", body, authHeader()...)

		if w.Code != http.StatusCreated {
			t.Fatalf("预期 201，实际得到 %d: %s", w.Code, w.Body.String())
		}

		// 验证用户已添加
		w2 := performRequest(r, "GET", "/api/v1/users", "", authHeader()...)
		var resp map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &resp)
		if resp["count"].(float64) != 3 {
			t.Errorf("预期 3 个用户，实际得到 %v", resp["count"])
		}
	})

	t.Run("无效数据 — 缺少必填字段", func(t *testing.T) {
		body := `{"username":""}`
		w := performRequest(r, "POST", "/api/v1/users", body, authHeader()...)
		if w.Code != http.StatusBadRequest {
			t.Errorf("预期 400，实际得到 %d", w.Code)
		}
	})

	t.Run("无效 JSON", func(t *testing.T) {
		body := `{invalid json}`
		w := performRequest(r, "POST", "/api/v1/users", body, authHeader()...)
		if w.Code != http.StatusBadRequest {
			t.Errorf("预期 400，实际得到 %d", w.Code)
		}
	})
}

// TestUpdateUser 测试更新用户
func TestUpdateUser(t *testing.T) {
	r := setupTestRouter()

	t.Run("更新存在的用户", func(t *testing.T) {
		body := `{"username":"alice_updated","email":"alice_new@example.com","age":29}`
		w := performRequest(r, "PUT", "/api/v1/users/1", body, authHeader()...)

		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d: %s", w.Code, w.Body.String())
		}

		// 验证更新结果
		w2 := performRequest(r, "GET", "/api/v1/users/1", "", authHeader()...)
		var user User
		json.Unmarshal(w2.Body.Bytes(), &user)
		if user.Username != "alice_updated" {
			t.Errorf("预期用户名 alice_updated，实际得到 %s", user.Username)
		}
	})

	t.Run("更新不存在的用户", func(t *testing.T) {
		body := `{"username":"nobody","email":"nobody@test.com","age":20}`
		w := performRequest(r, "PUT", "/api/v1/users/999", body, authHeader()...)
		if w.Code != http.StatusNotFound {
			t.Errorf("预期 404，实际得到 %d", w.Code)
		}
	})

	t.Run("更新请求体无效", func(t *testing.T) {
		w := performRequest(r, "PUT", "/api/v1/users/1", `{}`, authHeader()...)
		if w.Code != http.StatusBadRequest {
			t.Errorf("预期 400，实际得到 %d", w.Code)
		}
	})
}

// TestDeleteUser 测试删除用户
func TestDeleteUser(t *testing.T) {
	r := setupTestRouter()

	t.Run("删除存在的用户", func(t *testing.T) {
		w := performRequest(r, "DELETE", "/api/v1/users/1", "", authHeader()...)
		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", w.Code)
		}

		// 验证用户已删除
		w2 := performRequest(r, "GET", "/api/v1/users/1", "", authHeader()...)
		if w2.Code != http.StatusNotFound {
			t.Errorf("删除后获取用户应返回 404，实际得到 %d", w2.Code)
		}
	})

	t.Run("删除不存在的用户", func(t *testing.T) {
		w := performRequest(r, "DELETE", "/api/v1/users/999", "", authHeader()...)
		if w.Code != http.StatusNotFound {
			t.Errorf("预期 404，实际得到 %d", w.Code)
		}
	})
}

// TestSearchUsers 测试搜索用户
func TestSearchUsers(t *testing.T) {
	r := setupTestRouter()

	t.Run("按姓名搜索", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users/search?name=alice", "", authHeader()...)
		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["count"].(float64) != 1 {
			t.Errorf("预期 1 个结果，实际得到 %v", resp["count"])
		}
	})

	t.Run("按年龄过滤", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users/search?min_age=30", "", authHeader()...)
		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["count"].(float64) != 1 {
			t.Errorf("预期 1 个结果，实际得到 %v", resp["count"])
		}
	})

	t.Run("无匹配结果", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users/search?name=nobody", "", authHeader()...)
		if w.Code != http.StatusOK {
			t.Fatalf("预期 200，实际得到 %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["count"].(float64) != 0 {
			t.Errorf("预期 0 个结果，实际得到 %v", resp["count"])
		}
	})
}

// TestAuthMiddleware 测试认证中间件的各种场景
func TestAuthMiddleware(t *testing.T) {
	r := setupTestRouter()

	t.Run("缺少 Authorization 请求头", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users", "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("预期 401，实际得到 %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"] != "缺少 Authorization 请求头" {
			t.Errorf("错误信息不匹配: %v", resp["error"])
		}
	})

	t.Run("错误的 Bearer Token", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users", "", "Authorization", "Bearer wrong-token")
		if w.Code != http.StatusForbidden {
			t.Errorf("预期 403，实际得到 %d", w.Code)
		}
	})

	t.Run("正确的 Token", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/users", "", authHeader()...)
		if w.Code != http.StatusOK {
			t.Errorf("预期 200，实际得到 %d", w.Code)
		}
	})
}

// TestListArticles 测试文章列表接口（不需要认证）
func TestListArticles(t *testing.T) {
	r := setupTestRouter()
	w := performRequest(r, "GET", "/api/v1/articles", "")

	if w.Code != http.StatusOK {
		t.Fatalf("预期 200，实际得到 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 0 {
		t.Errorf("初始文章数为 0，实际得到 %v", resp["count"])
	}
}

// TestGetArticleByID 测试获取文章详情
func TestGetArticleByID(t *testing.T) {
	r := setupTestRouter()

	t.Run("文章不存在", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/articles/1", "")
		if w.Code != http.StatusNotFound {
			t.Errorf("预期 404，实际得到 %d", w.Code)
		}
	})

	t.Run("无效 ID", func(t *testing.T) {
		w := performRequest(r, "GET", "/api/v1/articles/xyz", "")
		if w.Code != http.StatusBadRequest {
			t.Errorf("预期 400，实际得到 %d", w.Code)
		}
	})
}

// TestConcurrentAccess 测试并发访问的安全性
func TestConcurrentAccess(t *testing.T) {
	r := setupTestRouter()

	// 并发创建用户和读取用户列表
	done := make(chan bool, 10)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			body := fmt.Sprintf(`{"username":"user%d","email":"user%d@test.com","age":%d}`, idx, idx, 20+idx)
			w := performRequest(r, "POST", "/api/v1/users", body, authHeader()...)
			if w.Code != http.StatusCreated {
				t.Errorf("并发创建用户失败: %d", w.Code)
			}
			done <- true
		}(i)
	}
	for i := 0; i < 5; i++ {
		go func() {
			w := performRequest(r, "GET", "/api/v1/users", "", authHeader()...)
			if w.Code != http.StatusOK {
				t.Errorf("并发读取用户列表失败: %d", w.Code)
			}
			done <- true
		}()
	}
	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证最终用户数
	w := performRequest(r, "GET", "/api/v1/users", "", authHeader()...)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	expected := float64(7) // 初始 2 个 + 新建 5 个
	if resp["count"].(float64) != expected {
		t.Errorf("预期 %v 个用户，实际得到 %v", expected, resp["count"])
	}
}