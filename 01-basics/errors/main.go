package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================
// 1. Error 接口
// ============================================================
// Go 的 error 是一个内置接口:
//
//	type error interface {
//	    Error() string
//	}
//
// 任何实现了 Error() string 方法的类型都可以作为 error 使用

// ============================================================
// 2. 哨兵错误 (Sentinel Errors)
// ============================================================
// 哨兵错误是预定义的包级别错误值，用于标识特定的错误条件
// 通常用 Err 前缀命名

var (
	ErrNotFound   = errors.New("资源未找到")
	ErrPermission = errors.New("权限不足")
	ErrTimeout    = errors.New("操作超时")
	ErrInvalidID  = errors.New("无效的 ID")
)

// ============================================================
// 3. 自定义错误类型
// ============================================================

// ValidationError 自定义错误类型 — 实现 error 接口
type ValidationError struct {
	Field   string // 出错的字段名
	Value   any    // 导致错误的原值
	Message string // 错误描述
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("验证错误: 字段=%q, 值=%v, 原因=%s",
		e.Field, e.Value, e.Message)
}

// Unwrap 可选方法：支持 errors.Is 和 errors.As 链式查找
func (e *ValidationError) Unwrap() error {
	return fmt.Errorf("字段 %s 验证失败", e.Field)
}

// NotFoundError 自定义错误
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s 未找到", e.Resource, e.ID)
}

// ============================================================
// 4. 带错误码的自定义错误
// ============================================================

type ErrorCode int

const (
	CodeInvalidInput   ErrorCode = 400
	CodeUnauthorized   ErrorCode = 401
	CodeNotFound       ErrorCode = 404
	CodeInternalError  ErrorCode = 500
	CodeServiceUnavail ErrorCode = 503
)

type AppError struct {
	Code    ErrorCode // 错误码
	Message string    // 人类可读的信息
	Err     error     // 原始错误（可选）
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误，支持 errors.Is 和 errors.As 链
func (e *AppError) Unwrap() error {
	return e.Err
}

// ============================================================
// 5. 实用函数：演示错误模式
// ============================================================

// 5a. 除法 — 基本错误返回
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为 0")
	}
	return a / b, nil
}

// 5b. 查找用户 — 哨兵错误 + 自定义错误
func findUser(id string) (string, error) {
	if id == "" {
		// 使用哨兵错误
		return "", fmt.Errorf("查找用户: %w", ErrInvalidID)
	}

	// 模拟数据库中没有找到
	if id != "u-001" {
		return "", &NotFoundError{
			Resource: "用户",
			ID:       id,
		}
	}
	return "张三", nil
}

// 5c. 创建用户 — 验证错误
func createUser(name string, age int) error {
	if name == "" {
		return &ValidationError{
			Field:   "name",
			Value:   name,
			Message: "用户名不能为空",
		}
	}
	if age < 0 || age > 150 {
		return &ValidationError{
			Field:   "age",
			Value:   age,
			Message: "年龄必须在 0-150 之间",
		}
	}
	if age < 18 {
		return &ValidationError{
			Field:   "age",
			Value:   age,
			Message: "未满 18 岁无法注册",
		}
	}
	fmt.Printf("  用户 %q (年龄 %d) 创建成功\n", name, age)
	return nil
}

// 5d. 读取配置文件 — 错误包装
func readConfig(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// 使用 %w 包装原始错误，保留错误链
		return "", fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	return string(data), nil
}

// 5e. 处理订单 — 多层错误包装
type OrderService struct{}

func (s *OrderService) ProcessOrder(orderID string) error {
	// 模拟各种错误情况
	if orderID == "" {
		return &AppError{
			Code:    CodeInvalidInput,
			Message: "订单 ID 不能为空",
		}
	}

	if orderID == "bad" {
		return &AppError{
			Code:    CodeInternalError,
			Message: "处理订单时发生内部错误",
			Err:     errors.New("数据库连接超时"),
		}
	}

	return nil // 成功
}

// ============================================================
// 6. main 函数 — 演示各种错误模式
// ============================================================

func main() {
	fmt.Println("=== Go 错误处理综合演示 ===")

	// ============================================================
	// 6a. 基本错误处理
	// ============================================================
	fmt.Println("--- 基本错误返回 ---")

	if result, err := divide(10, 3); err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("10 / 3 = %.2f\n", result)
	}

	if result, err := divide(10, 0); err != nil {
		fmt.Printf("10 / 0 = 错误: %v\n", err)
	} else {
		fmt.Printf("结果: %.2f\n", result)
	}

	// ============================================================
	// 6b. 哨兵错误 + errors.Is
	// ============================================================
	fmt.Println("\n--- 哨兵错误 + errors.Is ---")

	_, err := findUser("")
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			fmt.Printf("errors.Is 检测到 ErrInvalidID: %v\n", err)
		}
	}

	_, err = findUser("u-999")
	if err != nil {
		fmt.Printf("查找失败: %v\n", err)

		// errors.Is 可以沿着错误链查找
		var nf *NotFoundError
		if errors.As(err, &nf) {
			fmt.Printf("  errors.As 成功: Resource=%s, ID=%s\n", nf.Resource, nf.ID)
		}
	}

	// ============================================================
	// 6c. 自定义错误类型 + errors.As
	// ============================================================
	fmt.Println("\n--- 自定义错误类型 + errors.As ---")

	err = createUser("", 20)
	if err != nil {
		fmt.Printf("创建用户错误: %v\n", err)
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("  验证错误详情: 字段=%q, 值=%v, 原因=%s\n",
				ve.Field, ve.Value, ve.Message)
		}
	}

	err = createUser("小李", 15)
	if err != nil {
		fmt.Printf("创建用户错误: %v\n", err)
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("  验证错误详情: 字段=%q, 值=%v, 原因=%s\n",
				ve.Field, ve.Value, ve.Message)
		}
	}

	err = createUser("大王", 25)
	if err != nil {
		fmt.Printf("创建用户错误: %v\n", err)
	} else {
		fmt.Println("  大王创建成功!")
	}

	// ============================================================
	// 6d. fmt.Errorf + %w 包装
	// ============================================================
	fmt.Println("\n--- fmt.Errorf + %w 包装 ---")

	// 演示读取不存在的文件
	tmpFile := filepath.Join(os.TempDir(), "nonexistent-config.json")
	_, err = readConfig(tmpFile)
	if err != nil {
		fmt.Printf("错误 (已包装): %v\n", err)

		// 使用 errors.Is 可以穿透包装层，检查底层错误
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("  底层原因: 文件不存在 (通过 errors.Is 检测)")
		}
	}

	// 演示读取存在的文件
	goModPath := filepath.Join(".", "go.mod")
	content, err := readConfig(goModPath)
	if err != nil {
		fmt.Printf("读取错误: %v\n", err)
	} else {
		fmt.Printf("成功读取 go.mod (%d 字节):\n  %s\n",
			len(content), content)
	}

	// ============================================================
	// 6e. 多层错误包装实战
	// ============================================================
	fmt.Println("\n--- 多层错误包装实战 ---")

	svc := &OrderService{}

	// 场景 1: 空订单
	err = svc.ProcessOrder("")
	if err != nil {
		fmt.Printf("处理订单失败: %v\n", err)
		var appErr *AppError
		if errors.As(err, &appErr) {
			fmt.Printf("  错误码: %d\n", appErr.Code)
		}
	}

	// 场景 2: 内部错误
	err = svc.ProcessOrder("bad")
	if err != nil {
		fmt.Printf("处理订单失败: %v\n", err)
		var appErr *AppError
		if errors.As(err, &appErr) {
			fmt.Printf("  错误码: %d\n", appErr.Code)
			fmt.Printf("  原始错误: %v\n", appErr.Err)
			if appErr.Err != nil {
				fmt.Printf("  原始错误信息: %s\n", appErr.Err.Error())
			}
		}
	}

	// 场景 3: 正常处理
	err = svc.ProcessOrder("ORD-001")
	if err != nil {
		fmt.Printf("处理订单失败: %v\n", err)
	} else {
		fmt.Println("订单 ORD-001 处理成功 ✅")
	}

	// ============================================================
	// 6f. 常见错误模式对比
	// ============================================================
	fmt.Println("\n--- 常见错误模式对比 ---")

	fmt.Println("1. errors.New():")
	fmt.Println("   创建简单哨兵错误")
	fmt.Printf("   ErrNotFound = %v\n", ErrNotFound)

	fmt.Println("\n2. fmt.Errorf():")
	fmt.Println("   格式化错误消息，可用 %w 包装")
	err2 := fmt.Errorf("处理失败: 文件 %s %w", "config.yaml", os.ErrNotExist)
	fmt.Printf("   %v\n", err2)
	fmt.Printf("   errors.Is(err2, os.ErrNotExist) = %t\n",
		errors.Is(err2, os.ErrNotExist))

	fmt.Println("\n3. 自定义结构体:")
	fmt.Println("   携带额外上下文，支持 errors.As")

	fmt.Println("\n4. 错误对比总结:")
	fmt.Printf("   errors.Is(err, target) — 沿链查找相等错误\n")
	fmt.Printf("   errors.As(err, &target) — 沿链查找匹配类型\n")

	// ============================================================
	// 6g. 错误处理最佳实践
	// ============================================================
	fmt.Println("\n--- 错误处理最佳实践 ---")

	fmt.Println("✅ 总是检查错误 (不要用 _ 忽略)")
	fmt.Println("✅ 使用哨兵错误表示预期的错误条件")
	fmt.Println("✅ 使用 %w 包装错误，保留错误链")
	fmt.Println("✅ 使用 errors.Is 和 errors.As 检查错误")
	fmt.Println("✅ 自定义错误类型携带额外上下文")
	fmt.Println("✅ 错误消息小写开头（Go 约定）")
	fmt.Println("✅ 只在最外层处理错误（日志/用户提示）")
	fmt.Println("✅ 避免在错误消息中包含敏感信息")

	fmt.Println("\n❌ 不要在错误消息中使用大写开头或标点结尾")
	fmt.Println("   （例外: 专有名词和自定义格式）")
	fmt.Println("❌ 不要重复包装已包装过的错误")
}