package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Book 表示一本书
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at"`
}

// BookStore 内存中的书籍存储，带 JSON 持久化
type BookStore struct {
	mu     sync.RWMutex
	Books  []Book `json:"books"`
	NextID int    `json:"next_id"`
}

const dataFile = "books.json"

var store *BookStore

// loadStore 从 JSON 文件加载数据
func loadStore() (*BookStore, error) {
	s := &BookStore{Books: []Book{}, NextID: 1}
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil // 首次运行，返回空存储
		}
		return nil, err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("解析数据文件失败: %w", err)
	}
	return s, nil
}

// save 将存储持久化到 JSON 文件
func (s *BookStore) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0644)
}

// --- 辅助函数 ---

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// getIDFromPath 从路径中提取 ID，例如 /books/42 -> 42
func getIDFromPath(path string) (int, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/books/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, false
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return id, true
}

// --- 路由分发 ---

// booksHandler 处理 /books 路径下的所有请求
func booksHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /books 路径（无 ID）
	if path == "/books" || path == "/books/" {
		switch r.Method {
		case http.MethodGet:
			listBooks(w, r)
		case http.MethodPost:
			createBook(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "不支持的 HTTP 方法")
		}
		return
	}

	// /books/{id} 路径
	id, ok := getIDFromPath(path)
	if !ok {
		writeError(w, http.StatusBadRequest, "无效的书籍 ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		getBook(w, r, id)
	case http.MethodPut:
		updateBook(w, r, id)
	case http.MethodDelete:
		deleteBook(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的 HTTP 方法")
	}
}

// --- CRUD 处理器 ---

// listBooks GET /books - 列出所有书籍，支持查询参数过滤和分页
func listBooks(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	books := store.Books

	// 查询参数过滤
	q := r.URL.Query()

	// 按标题搜索
	if title := q.Get("title"); title != "" {
		var filtered []Book
		for _, b := range books {
			if strings.Contains(strings.ToLower(b.Title), strings.ToLower(title)) {
				filtered = append(filtered, b)
			}
		}
		books = filtered
	}

	// 按作者搜索
	if author := q.Get("author"); author != "" {
		var filtered []Book
		for _, b := range books {
			if strings.Contains(strings.ToLower(b.Author), strings.ToLower(author)) {
				filtered = append(filtered, b)
			}
		}
		books = filtered
	}

	// 按年份过滤
	if yearStr := q.Get("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err == nil {
			var filtered []Book
			for _, b := range books {
				if b.Year == year {
					filtered = append(filtered, b)
				}
			}
			books = filtered
		}
	}

	// 分页
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	if start >= len(books) {
		books = []Book{}
	} else {
		end := start + pageSize
		if end > len(books) {
			end = len(books)
		}
		books = books[start:end]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"books":      books,
		"page":       page,
		"page_size":  pageSize,
		"total":      len(store.Books),
		"filtered":   len(books),
	})
}

// createBook POST /books - 创建新书
func createBook(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		Year   int    `json:"year"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	// 基本校验
	if input.Title == "" || input.Author == "" {
		writeError(w, http.StatusBadRequest, "标题和作者不能为空")
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	book := Book{
		ID:        store.NextID,
		Title:     input.Title,
		Author:    input.Author,
		Year:      input.Year,
		CreatedAt: time.Now(),
	}
	store.Books = append(store.Books, book)
	store.NextID++

	// 持久化
	if err := store.save(); err != nil {
		log.Printf("持久化失败: %v", err)
	}

	writeJSON(w, http.StatusCreated, book)
}

// getBook GET /books/{id} - 获取单本书
func getBook(w http.ResponseWriter, r *http.Request, id int) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, b := range store.Books {
		if b.ID == id {
			writeJSON(w, http.StatusOK, b)
			return
		}
	}
	writeError(w, http.StatusNotFound, fmt.Sprintf("未找到 ID 为 %d 的书籍", id))
}

// updateBook PUT /books/{id} - 更新书籍
func updateBook(w http.ResponseWriter, r *http.Request, id int) {
	var input struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		Year   int    `json:"year"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for i := range store.Books {
		if store.Books[i].ID == id {
			if input.Title != "" {
				store.Books[i].Title = input.Title
			}
			if input.Author != "" {
				store.Books[i].Author = input.Author
			}
			if input.Year > 0 {
				store.Books[i].Year = input.Year
			}
			if err := store.save(); err != nil {
				log.Printf("持久化失败: %v", err)
			}
			writeJSON(w, http.StatusOK, store.Books[i])
			return
		}
	}
	writeError(w, http.StatusNotFound, fmt.Sprintf("未找到 ID 为 %d 的书籍", id))
}

// deleteBook DELETE /books/{id} - 删除书籍
func deleteBook(w http.ResponseWriter, r *http.Request, id int) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for i := range store.Books {
		if store.Books[i].ID == id {
			deleted := store.Books[i]
			store.Books = append(store.Books[:i], store.Books[i+1:]...)
			if err := store.save(); err != nil {
				log.Printf("持久化失败: %v", err)
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"message": "删除成功",
				"book":    deleted,
			})
			return
		}
	}
	writeError(w, http.StatusNotFound, fmt.Sprintf("未找到 ID 为 %d 的书籍", id))
}

// --- 日志中间件 ---

// loggingMiddleware 记录每个请求的方法、路径和耗时
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	// 初始化存储
	var err error
	store, err = loadStore()
	if err != nil {
		log.Fatalf("加载存储失败: %v", err)
	}
	log.Printf("已加载 %d 本书籍，下一 ID: %d", len(store.Books), store.NextID)

	// 注册路由
	mux := http.NewServeMux()
	mux.HandleFunc("/books", booksHandler)
	mux.HandleFunc("/books/", booksHandler)

	// 根路径提示
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeError(w, http.StatusNotFound, "路径不存在")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"name":    "书籍管理 API",
			"version": "1.0.0",
			"endpoints": "GET /books, POST /books, GET /books/{id}, PUT /books/{id}, DELETE /books/{id}",
		})
	})

	// 包装日志中间件
	handler := loggingMiddleware(mux)

	addr := ":8080"
	log.Printf("📚 书籍管理 API 启动在 http://localhost%s", addr)
	log.Printf("📖 示例: curl http://localhost%s/books", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}