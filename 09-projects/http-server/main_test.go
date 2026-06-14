package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetIDFromPath 测试从 URL 路径中提取 ID
func TestGetIDFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantID   int
		wantOK   bool
	}{
		{name: "标准路径 /books/42", path: "/books/42", wantID: 42, wantOK: true},
		{name: "带额外斜杠", path: "/books/99/", wantID: 99, wantOK: true},
		{name: "无 ID", path: "/books/", wantID: 0, wantOK: false},
		{name: "根路径", path: "/books", wantID: 0, wantOK: false},
		{name: "非数字 ID", path: "/books/abc", wantID: 0, wantOK: false},
		{name: "空路径", path: "", wantID: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := getIDFromPath(tt.path)
			if id != tt.wantID {
				t.Errorf("getIDFromPath(%q) id = %d, 期望 %d", tt.path, id, tt.wantID)
			}
			if ok != tt.wantOK {
				t.Errorf("getIDFromPath(%q) ok = %v, 期望 %v", tt.path, ok, tt.wantOK)
			}
		})
	}
}

// TestWriteJSON 测试 writeJSON 辅助函数
func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}

	writeJSON(recorder, http.StatusOK, data)

	if recorder.Code != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 %d", recorder.Code, http.StatusOK)
	}
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, 期望 %q", contentType, "application/json; charset=utf-8")
	}

	var decoded map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON 解码失败: %v", err)
	}
	if decoded["hello"] != "world" {
		t.Errorf("body = %v, 期望 hello=world", decoded)
	}
}

// TestWriteError 测试 writeError 辅助函数
func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeError(recorder, http.StatusBadRequest, "参数错误")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, 期望 %d", recorder.Code, http.StatusBadRequest)
	}

	var body map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &body)
	if body["error"] != "参数错误" {
		t.Errorf("error = %q, 期望 %q", body["error"], "参数错误")
	}
}

// setupTestStore 初始化测试用 BookStore
func setupTestStore() {
	store = &BookStore{
		Books: []Book{
			{ID: 1, Title: "Go 编程", Author: "作者A", Year: 2020},
			{ID: 2, Title: "分布式系统", Author: "作者B", Year: 2021},
			{ID: 3, Title: "算法导论", Author: "作者C", Year: 2019},
		},
		NextID: 4,
	}
}

// TestListBooks 测试 GET /books
func TestListBooks(t *testing.T) {
	setupTestStore()

	t.Run("列出所有书籍", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("状态码 = %d, 期望 %d", rec.Code, http.StatusOK)
		}

		var resp map[string]any
		json.Unmarshal(rec.Body.Bytes(), &resp)

		total := resp["total"].(float64)
		if total != 3 {
			t.Errorf("total = %.0f, 期望 3", total)
		}
	})

	t.Run("按标题搜索", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books?title=Go", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		var resp map[string]any
		json.Unmarshal(rec.Body.Bytes(), &resp)

		books := resp["books"].([]any)
		if len(books) != 1 {
			t.Fatalf("搜索 title=Go 期望 1 个结果，得到 %d", len(books))
		}
	})

	t.Run("按作者搜索", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books?author=作者B", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		var resp map[string]any
		json.Unmarshal(rec.Body.Bytes(), &resp)

		books := resp["books"].([]any)
		if len(books) != 1 {
			t.Fatalf("搜索 author=作者B 期望 1 个结果，得到 %d", len(books))
		}
	})

	t.Run("按年份过滤", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books?year=2020", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		var resp map[string]any
		json.Unmarshal(rec.Body.Bytes(), &resp)

		books := resp["books"].([]any)
		if len(books) != 1 {
			t.Fatalf("过滤 year=2020 期望 1 个结果，得到 %d", len(books))
		}
	})

	t.Run("分页测试", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books?page=1&page_size=2", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		var resp map[string]any
		json.Unmarshal(rec.Body.Bytes(), &resp)

		books := resp["books"].([]any)
		if len(books) != 2 {
			t.Fatalf("分页 page_size=2 期望 2 个结果，得到 %d", len(books))
		}
		if resp["page"].(float64) != 1 {
			t.Errorf("page 期望 1，得到 %.0f", resp["page"].(float64))
		}
		if resp["page_size"].(float64) != 2 {
			t.Errorf("page_size 期望 2，得到 %.0f", resp["page_size"].(float64))
		}
	})
}

// TestCreateBook 测试 POST /books
func TestCreateBook(t *testing.T) {
	setupTestStore()

	t.Run("创建书籍成功", func(t *testing.T) {
		body := `{"title":"新书","author":"新作者","year":2023}`
		req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("状态码 = %d, 期望 %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var book Book
		json.Unmarshal(rec.Body.Bytes(), &book)
		if book.ID != 4 {
			t.Errorf("ID 期望 4，得到 %d", book.ID)
		}
		if book.Title != "新书" {
			t.Errorf("Title 期望 '新书'，得到 '%s'", book.Title)
		}
	})

	t.Run("标题为空返回 400", func(t *testing.T) {
		body := `{"title":"","author":"作者","year":2023}`
		req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("请求体格式错误返回 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusBadRequest)
		}
	})
}

// TestGetBook 测试 GET /books/{id}
func TestGetBook(t *testing.T) {
	setupTestStore()

	t.Run("获取存在的书籍", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("状态码 = %d, 期望 %d", rec.Code, http.StatusOK)
		}

		var book Book
		json.Unmarshal(rec.Body.Bytes(), &book)
		if book.Title != "Go 编程" {
			t.Errorf("Title 期望 'Go 编程'，得到 '%s'", book.Title)
		}
	})

	t.Run("获取不存在的书籍返回 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books/999", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("无效 ID 返回 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/books/abc", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusBadRequest)
		}
	})
}

// TestUpdateBook 测试 PUT /books/{id}
func TestUpdateBook(t *testing.T) {
	setupTestStore()

	t.Run("更新书籍成功", func(t *testing.T) {
		body := `{"title":"Go 编程（第2版）","year":2024}`
		req := httptest.NewRequest(http.MethodPut, "/books/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("状态码 = %d, 期望 %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var book Book
		json.Unmarshal(rec.Body.Bytes(), &book)
		if book.Title != "Go 编程（第2版）" {
			t.Errorf("Title 期望 'Go 编程（第2版）'，得到 '%s'", book.Title)
		}
		// 确认作者未被修改
		if book.Author != "作者A" {
			t.Errorf("Author 期望 '作者A'，得到 '%s'", book.Author)
		}
	})

	t.Run("更新不存在的书籍返回 404", func(t *testing.T) {
		body := `{"title":"不存在"}`
		req := httptest.NewRequest(http.MethodPut, "/books/999", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusNotFound)
		}
	})
}

// TestDeleteBook 测试 DELETE /books/{id}
func TestDeleteBook(t *testing.T) {
	setupTestStore()

	t.Run("删除存在的书籍", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/books/3", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("状态码 = %d, 期望 %d", rec.Code, http.StatusOK)
		}

		var resp map[string]any
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["message"] != "删除成功" {
			t.Errorf("message = %v, 期望 '删除成功'", resp["message"])
		}
	})

	t.Run("删除不存在的书籍返回 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/books/999", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusNotFound)
		}
	})
}

// TestMethodNotAllowed 测试不支持的 HTTP 方法
func TestMethodNotAllowed(t *testing.T) {
	setupTestStore()

	t.Run("PATCH 方法返回 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/books", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("PATCH 单个书籍返回 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/books/1", nil)
		rec := httptest.NewRecorder()
		booksHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}

// TestRootEndpoint 测试根路径
func TestRootEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// 模拟 main 中注册的路由
	mux := http.NewServeMux()
	mux.HandleFunc("/books", booksHandler)
	mux.HandleFunc("/books/", booksHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeError(w, http.StatusNotFound, "路径不存在")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"name":    "书籍管理 API",
			"version": "1.0.0",
		})
	})
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusOK)
	}
}

// TestNotFound 测试不存在的路径
func TestNotFound(t *testing.T) {
	setupTestStore()
	// 重置 store 以确保测试独立

	req := httptest.NewRequest(http.MethodGet, "/nonexist", nil)
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/books", booksHandler)
	mux.HandleFunc("/books/", booksHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeError(w, http.StatusNotFound, "路径不存在")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"name": "test"})
	})
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusNotFound)
	}
}