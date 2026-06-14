// main_test.go — 对 errors 包中哨兵错误、自定义错误类型、错误包装等特性的完整测试套件
//
// 本文件演示 Go 测试的多种模式：
//   - 表格驱动测试（Table-Driven Tests）
//   - 子测试（t.Run）
//   - 哨兵错误断言测试（errors.Is）
//   - 自定义错误类型断言测试（errors.As）
//   - 错误包装链测试
//   - 边界值测试
package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// 1. 哨兵错误测试
// ============================================================

// TestSentinelErrors 验证哨兵错误定义
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "ErrNotFound", err: ErrNotFound, want: "资源未找到"},
		{name: "ErrPermission", err: ErrPermission, want: "权限不足"},
		{name: "ErrTimeout", err: ErrTimeout, want: "操作超时"},
		{name: "ErrInvalidID", err: ErrInvalidID, want: "无效的 ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("哨兵错误消息 = %q; 期望 %q", tt.err.Error(), tt.want)
			}
		})
	}
}

// ============================================================
// 2. 基本错误返回测试
// ============================================================

// TestDivide 表格驱动测试带错误返回的除法
func TestDivide(t *testing.T) {
	t.Run("正常除法", func(t *testing.T) {
		got, err := divide(10, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 10.0/3.0 {
			t.Errorf("divide(10, 3) = %f; 期望 %f", got, 10.0/3.0)
		}
	})

	t.Run("整除", func(t *testing.T) {
		got, err := divide(9, 3)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 3.0 {
			t.Errorf("divide(9, 3) = %f; 期望 3.0", got)
		}
	})

	t.Run("除零错误", func(t *testing.T) {
		_, err := divide(10, 0)
		if err == nil {
			t.Fatal("期望除零错误，但没有得到")
		}
		if err.Error() != "除数不能为 0" {
			t.Errorf("错误消息 = %q; 期望 %q", err.Error(), "除数不能为 0")
		}
	})

	t.Run("零除以正数", func(t *testing.T) {
		got, err := divide(0, 5)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if got != 0 {
			t.Errorf("divide(0, 5) = %f; 期望 0", got)
		}
	})
}

// ============================================================
// 3. 哨兵错误 + errors.Is 测试
// ============================================================

// TestFindUserErrors 测试 findUser 的错误返回
func TestFindUserErrors(t *testing.T) {
	t.Run("空 ID 返回 ErrInvalidID", func(t *testing.T) {
		_, err := findUser("")
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		if !errors.Is(err, ErrInvalidID) {
			t.Errorf("期望 errors.Is(err, ErrInvalidID) 为真，得到 %v", err)
		}
	})

	t.Run("不存在用户返回 NotFoundError", func(t *testing.T) {
		_, err := findUser("u-999")
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var nf *NotFoundError
		if !errors.As(err, &nf) {
			t.Fatalf("期望 NotFoundError 类型，得到 %T", err)
		}
		if nf.Resource != "用户" {
			t.Errorf("NotFoundError.Resource = %q; 期望 %q", nf.Resource, "用户")
		}
		if nf.ID != "u-999" {
			t.Errorf("NotFoundError.ID = %q; 期望 %q", nf.ID, "u-999")
		}
	})

	t.Run("正确 ID 返回用户", func(t *testing.T) {
		user, err := findUser("u-001")
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if user != "张三" {
			t.Errorf("findUser 返回 %q; 期望 %q", user, "张三")
		}
	})
}

// ============================================================
// 4. 自定义错误类型 + errors.As 测试
// ============================================================

// TestCreateUser 表格驱动测试 createUser 的验证错误
func TestCreateUser(t *testing.T) {
	t.Run("空用户名返回 ValidationError", func(t *testing.T) {
		err := createUser("", 20)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("期望 ValidationError 类型，得到 %T", err)
		}
		if ve.Field != "name" {
			t.Errorf("Field = %q; 期望 %q", ve.Field, "name")
		}
		if ve.Message != "用户名不能为空" {
			t.Errorf("Message = %q; 期望 %q", ve.Message, "用户名不能为空")
		}
	})

	t.Run("年龄为负返回 ValidationError", func(t *testing.T) {
		err := createUser("小李", -5)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("期望 ValidationError 类型，得到 %T", err)
		}
		if ve.Field != "age" {
			t.Errorf("Field = %q; 期望 %q", ve.Field, "age")
		}
	})

	t.Run("年龄超过 150 返回 ValidationError", func(t *testing.T) {
		err := createUser("小李", 200)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("期望 ValidationError 类型，得到 %T", err)
		}
	})

	t.Run("未满 18 岁返回 ValidationError", func(t *testing.T) {
		err := createUser("小李", 15)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("期望 ValidationError 类型，得到 %T", err)
		}
		if ve.Message != "未满 18 岁无法注册" {
			t.Errorf("Message = %q; 期望 %q", ve.Message, "未满 18 岁无法注册")
		}
	})

	t.Run("合法用户创建成功", func(t *testing.T) {
		err := createUser("大王", 25)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
	})

	t.Run("边界值：18 岁创建成功", func(t *testing.T) {
		err := createUser("小王", 18)
		if err != nil {
			t.Fatalf("18 岁应能注册，得到 %v", err)
		}
	})
}

// ============================================================
// 5. 错误包装测试（fmt.Errorf + %w）
// ============================================================

// TestReadConfig 测试错误包装
func TestReadConfig(t *testing.T) {
	t.Run("不存在的文件返回包装错误", func(t *testing.T) {
		tmpFile := filepath.Join(os.TempDir(), "nonexistent-test-file-12345.json")
		_, err := readConfig(tmpFile)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		// 验证包装后的错误包含上下文
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("期望 errors.Is(err, os.ErrNotExist) 为真")
		}
		// 验证消息包含了路径信息
		if !contains(err.Error(), tmpFile) {
			t.Errorf("错误消息应包含路径 %q", tmpFile)
		}
	})
}

// TestReadConfigGoMod 测试读取存在的文件（使用项目根目录的 go.mod）
func TestReadConfigGoMod(t *testing.T) {
	goModPath := filepath.Join("..", "..", "go.mod")
	content, err := readConfig(goModPath)
	if err != nil {
		t.Fatalf("读取 go.mod 失败: %v", err)
	}
	if len(content) == 0 {
		t.Error("go.mod 内容不应为空")
	}
}

// ============================================================
// 6. 多层错误包装测试
// ============================================================

// TestOrderService 测试 OrderService 的错误处理
func TestOrderService(t *testing.T) {
	svc := &OrderService{}

	t.Run("空订单 ID 返回 AppError", func(t *testing.T) {
		err := svc.ProcessOrder("")
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var appErr *AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("期望 AppError 类型，得到 %T", err)
		}
		if appErr.Code != CodeInvalidInput {
			t.Errorf("错误码 = %d; 期望 %d (CodeInvalidInput)", appErr.Code, CodeInvalidInput)
		}
	})

	t.Run("内部错误包含原始错误", func(t *testing.T) {
		err := svc.ProcessOrder("bad")
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		var appErr *AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("期望 AppError 类型，得到 %T", err)
		}
		if appErr.Code != CodeInternalError {
			t.Errorf("错误码 = %d; 期望 %d (CodeInternalError)", appErr.Code, CodeInternalError)
		}
		if appErr.Err == nil {
			t.Fatal("期望 AppError.Err 不为 nil")
		}
		if appErr.Err.Error() != "数据库连接超时" {
			t.Errorf("原始错误消息 = %q; 期望 %q", appErr.Err.Error(), "数据库连接超时")
		}
	})

	t.Run("正常订单处理成功", func(t *testing.T) {
		err := svc.ProcessOrder("ORD-001")
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
	})
}

// ============================================================
// 7. 错误码测试
// ============================================================

// TestErrorCodes 验证错误码常量
func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
		want ErrorCode
	}{
		{name: "CodeInvalidInput", code: CodeInvalidInput, want: 400},
		{name: "CodeUnauthorized", code: CodeUnauthorized, want: 401},
		{name: "CodeNotFound", code: CodeNotFound, want: 404},
		{name: "CodeInternalError", code: CodeInternalError, want: 500},
		{name: "CodeServiceUnavail", code: CodeServiceUnavail, want: 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("错误码 = %d; 期望 %d", tt.code, tt.want)
			}
		})
	}
}

// ============================================================
// 8. AppError 格式化测试
// ============================================================

// TestAppErrorMessage 验证 AppError 的错误消息格式
func TestAppErrorMessage(t *testing.T) {
	t.Run("无原始错误", func(t *testing.T) {
		err := &AppError{Code: CodeInvalidInput, Message: "订单 ID 不能为空"}
		want := "[400] 订单 ID 不能为空"
		if err.Error() != want {
			t.Errorf("AppError.Error() = %q; 期望 %q", err.Error(), want)
		}
	})

	t.Run("有原始错误", func(t *testing.T) {
		err := &AppError{
			Code:    CodeInternalError,
			Message: "处理订单时发生内部错误",
			Err:     errors.New("数据库连接超时"),
		}
		want := "[500] 处理订单时发生内部错误: 数据库连接超时"
		if err.Error() != want {
			t.Errorf("AppError.Error() = %q; 期望 %q", err.Error(), want)
		}
	})
}

// ============================================================
// 辅助函数
// ============================================================

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstring(s, substr)
}

// containsSubstring 简单子串查找
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}