// Package calculator 提供基本的四则运算功能，是 Go 测试模式的演示示例。
//
// 本文件展示了如何构建一个可测试的包，包含：
//   - 标准函数定义（Add, Subtract, Multiply, Divide）
//   - 错误返回模式（Divide 返回 error）
//   - 导出的类型化错误变量
package calculator

import (
	"errors"
	"fmt"
)

// ErrDivideByZero 是除零错误的哨兵错误（sentinel error）
var ErrDivideByZero = errors.New("除数不能为零")

// Add 返回两个整数的和
func Add(a, b int) int {
	return a + b
}

// Subtract 返回两个整数的差（a - b）
func Subtract(a, b int) int {
	return a - b
}

// Multiply 返回两个整数的积
func Multiply(a, b int) int {
	return a * b
}

// Divide 返回两个整数的商。
// 如果除数为零，返回 0 和 ErrDivideByZero 错误。
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("%w: a=%d, b=%d", ErrDivideByZero, a, b)
	}
	return a / b, nil
}