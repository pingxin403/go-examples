// main.go - Gin Web 框架示例
//
// 本示例演示 Gin 的核心功能：
//   - 基本路由 (GET/POST/PUT/DELETE)
//   - 路径参数与查询参数
//   - JSON 绑定与验证
//   - 中间件（日志记录）
//   - 路由分组
//   - 静态文件服务
//   - HTML 渲染
//
// 安装依赖:
//   go get github.com/gin-gonic/gin

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------- 模型定义 ----------

// User 用户模型，用于 JSON 绑定和验证
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"gte=0,lte=150"`
}

// Article 文章模型
type Article struct {
	ID      uint   `json:"id"`
	Title   string `json:"title" binding:"required,min=1"`
	Content string `json:"content" binding:"required"`
}

// ---------- 模拟数据存储 ----------

var (
	users    = map[uint]User{}
	articles = map[uint]Article{}
	nextID   = uint(1)
	mu       sync.Mutex // 保护 users/articles/nextID 的并发访问
)

func init() {
	// 初始化示例数据
	users[1] = User{ID: 1, Username: "alice", Email: "alice@example.com", Age: 28}
	users[2] = User{ID: 2, Username: "bob", Email: "bob@example.com", Age: 35}
	nextID = 3
}

// ---------- 中间件 ----------

// LoggerMiddleware 自定义日志中间件，记录每个请求的方法、路径、状态码和耗时
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start)
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		log.Printf("[%s] %s %d | %v", method, path, status, duration)
	}
}

// AuthMiddleware 模拟认证中间件，检查请求头中的 Authorization
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization 请求头"})
			c.Abort() // 终止请求链
			return
		}
		if token != "Bearer my-secret-token" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无效的 Token"})
			c.Abort()
			return
		}
		// 将用户信息存入上下文，后续处理函数可以读取
		c.Set("current_user", "admin")
		c.Next()
	}
}

// ---------- 处理函数 ----------

// listUsers 获取所有用户 (GET /api/users)
func listUsers(c *gin.Context) {
	mu.Lock()
	var userList []User
	for _, u := range users {
		userList = append(userList, u)
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(userList),
		"data":  userList,
	})
	mu.Unlock()
}

// getUserByID 根据 ID 获取单个用户 (GET /api/users/:id)
func getUserByID(c *gin.Context) {
	id := c.Param("id")
	var uid uint
	if _, err := fmt.Sscanf(id, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	mu.Lock()
	user, ok := users[uid]
	mu.Unlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// createUser 创建新用户 (POST /api/users)
func createUser(c *gin.Context) {
	var newUser User

	// 绑定 JSON 请求体并验证
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据无效",
			"details": err.Error(),
		})
		return
	}

	mu.Lock()
	newUser.ID = nextID
	nextID++
	users[newUser.ID] = newUser
	mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"data":    newUser,
	})
}

// updateUser 更新用户信息 (PUT /api/users/:id)
func updateUser(c *gin.Context) {
	id := c.Param("id")
	var uid uint
	if _, err := fmt.Sscanf(id, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	mu.Lock()
	if _, ok := users[uid]; !ok {
		mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	mu.Unlock()

	var updatedUser User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求数据无效",
			"details": err.Error(),
		})
		return
	}

	mu.Lock()
	updatedUser.ID = uid
	users[uid] = updatedUser
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
		"data":    updatedUser,
	})
}

// deleteUser 删除用户 (DELETE /api/users/:id)
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	var uid uint
	if _, err := fmt.Sscanf(id, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	mu.Lock()
	if _, ok := users[uid]; !ok {
		mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	delete(users, uid)
	mu.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("用户 %d 已删除", uid),
	})
}

// searchUsers 搜索用户，演示查询参数 (GET /api/users/search?name=xxx&min_age=18)
func searchUsers(c *gin.Context) {
	name := c.Query("name")
	minAgeStr := c.DefaultQuery("min_age", "0")

	var minAge int
	fmt.Sscanf(minAgeStr, "%d", &minAge)

	mu.Lock()
	var results []User
	for _, u := range users {
		if name != "" && u.Username != name {
			continue
		}
		if u.Age < minAge {
			continue
		}
		results = append(results, u)
	}
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"count": len(results),
		"data":  results,
	})
}

// ---------- HTML 渲染 ----------

func renderHome(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "Gin 示例",
		"message": "欢迎来到 Gin Web 框架示例！",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// ---------- 查询参数演示 ----------

func greetUser(c *gin.Context) {
	name := c.DefaultQuery("name", "游客")
	c.String(http.StatusOK, "你好，%s！", name)
}

// ---------- 主函数 ----------

func main() {
	// =========================================================
	// 1. 创建 Gin 引擎（默认包含 Logger 和 Recovery 中间件）
	// =========================================================
	r := gin.Default()

	// =========================================================
	// 2. 创建日志文件，将 Gin 日志写入文件而非控制台
	// =========================================================
	logFile, err := os.Create("gin.log")
	if err != nil {
		log.Fatalf("创建日志文件失败: %v", err)
	}
	defer logFile.Close()
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	// =========================================================
	// 3. 加载 HTML 模板（使用 Go 标准库模板引擎）
	// =========================================================
	r.LoadHTMLGlob("templates/*")

	// =========================================================
	// 4. 静态文件服务 — 提供 /static 路径下的静态资源
	// =========================================================
	r.Static("/static", "./static")

	// =========================================================
	// 5. 基本路由 — 最简单的 GET/POST/PUT/DELETE
	// =========================================================
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/hello", greetUser)

	r.GET("/", renderHome)

	// =========================================================
	// 6. 路由分组 — API v1
	// =========================================================
	v1 := r.Group("/api/v1")
	{
		// 文章相关接口（公开）
		articles := v1.Group("/articles")
		{
			articles.GET("", listArticles)
			articles.GET("/:id", getArticleByID)
		}

		// 用户相关接口（需要认证）
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

	// =========================================================
	// 7. 路由分组 — 带中间件的管理后台
	// =========================================================
	admin := r.Group("/admin")
	admin.Use(LoggerMiddleware(), AuthMiddleware())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("current_user")
			c.JSON(http.StatusOK, gin.H{
				"message": "管理后台首页",
				"user":    user,
				"stats": gin.H{
					"users":    len(users),
					"articles": len(articles),
				},
			})
		})
	}

	// =========================================================
	// 8. 启动服务器
	// =========================================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("启动服务器，监听端口 %s", port)
	log.Printf("可用接口:")
	log.Printf("  GET  /ping                    — 健康检查")
	log.Printf("  GET  /hello?name=xxx          — 问候")
	log.Printf("  GET  /                        — 首页 HTML")
	log.Printf("  GET  /api/v1/articles          — 文章列表")
	log.Printf("  GET  /api/v1/articles/:id     — 文章详情")
	log.Printf("  GET  /api/v1/users             — 用户列表（需认证）")
	log.Printf("  GET  /api/v1/users/:id        — 用户详情（需认证）")
	log.Printf("  GET  /api/v1/users/search     — 搜索用户（需认证）")
	log.Printf("  POST /api/v1/users             — 创建用户（需认证）")
	log.Printf("  PUT  /api/v1/users/:id        — 更新用户（需认证）")
	log.Printf("  DEL  /api/v1/users/:id        — 删除用户（需认证）")
	log.Printf("  GET  /admin/dashboard          — 管理后台（需认证）")

	_ = r.Run(":" + port)
}

// ---------- Article 模拟处理函数 ----------

func listArticles(c *gin.Context) {
	var list []Article
	for _, a := range articles {
		list = append(list, a)
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(list),
		"data":  list,
	})
}

func getArticleByID(c *gin.Context) {
	id := c.Param("id")
	var aid uint
	if _, err := fmt.Sscanf(id, "%d", &aid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章 ID"})
		return
	}
	article, ok := articles[aid]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	c.JSON(http.StatusOK, article)
}