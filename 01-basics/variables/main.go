package main

import (
	"fmt"
	"math"
	"unicode"
)

// 包级别常量 — 未导出（小写开头），仅在包内可见
const unexportedConst = "我只能在本包内使用"

// Pi 导出常量（大写开头），其他包可以引用
const Pi = 3.1415926

// 使用 iota 定义枚举常量
const (
	// iota 从 0 开始，每行递增
	StatusPending = iota // 0
	StatusActive         // 1
	StatusInactive       // 2
	StatusDeleted        // 3
)

// 使用 iota 做位运算
const (
	Read  = 1 << iota // 1 (0001)
	Write             // 2 (0010)
	Execute           // 4 (0100)
)

// 包级别的零值演示
var (
	zeroInt    int
	zeroFloat  float64
	zeroBool   bool
	zeroString string
	zeroPtr    *int
	zeroSlice  []int
	zeroMap    map[string]int
)

func main() {
	// ============================================================
	// 1. 短声明 :=  — 最常用的声明方式，类型由编译器推断
	// ============================================================
	fmt.Println("=== 短声明 := ===")
	name := "张三"       // string
	age := 30           // int
	height := 1.75      // float64
	active := true      // bool
	fmt.Printf("name=%s (%T), age=%d (%T), height=%.2f (%T), active=%t (%T)\n",
		name, name, age, age, height, height, active, active)

	// 短声明可以重新声明（只要至少有一个新变量）
	count, err := 10, error(nil) // 首次声明——假装有 err 只是为了演示
	fmt.Printf("count=%d, err=%v\n", count, err)

	// ============================================================
	// 2. var 声明 — 显式类型，适合零值初始化或包级别
	// ============================================================
	fmt.Println("\n=== var 声明 ===")
	var city string = "北京" // 完整声明
	var country = "中国"     // 省略类型，编译器推断
	var population int      // 不赋值 → 零值: 0
	fmt.Printf("city=%s, country=%s, population=%d\n", city, country, population)

	// 批量 var 声明
	var (
		firstName = "李"
		lastName  = "四"
		score     = 95.5
	)
	fmt.Printf("firstName=%s, lastName=%s, score=%.1f\n", firstName, lastName, score)

	// ============================================================
	// 3. 多重赋值
	// ============================================================
	fmt.Println("\n=== 多重赋值 ===")
	x, y := 1, 2
	fmt.Printf("交换前: x=%d, y=%d\n", x, y)
	x, y = y, x // 交换变量 — 不需要临时变量
	fmt.Printf("交换后: x=%d, y=%d\n", x, y)

	// 函数返回多个值
	quotient, remainder := divide(10, 3)
	fmt.Printf("10 / 3 = %d 余 %d\n", quotient, remainder)

	// ============================================================
	// 4. 零值（Zero Values）— Go 变量不初始化也有默认值
	// ============================================================
	fmt.Println("\n=== 零值（Zero Values）===")
	fmt.Printf("int:     %d\n", zeroInt)
	fmt.Printf("float64: %f\n", zeroFloat)
	fmt.Printf("bool:    %t\n", zeroBool)
	fmt.Printf("string:  %q\n", zeroString)   // 空字符串
	fmt.Printf("pointer: %v (nil=%t)\n", zeroPtr, zeroPtr == nil)
	fmt.Printf("slice:   %v (nil=%t, len=%d)\n", zeroSlice, zeroSlice == nil, len(zeroSlice))
	fmt.Printf("map:     %v (nil=%t)\n", zeroMap, zeroMap == nil)

	// ============================================================
	// 5. 常量与 iota
	// ============================================================
	fmt.Println("\n=== 常量与 iota ===")
	fmt.Printf("Pi=%.4f\n", Pi)

	fmt.Println("订单状态枚举:")
	fmt.Printf("  StatusPending  = %d (待处理)\n", StatusPending)
	fmt.Printf("  StatusActive   = %d (进行中)\n", StatusActive)
	fmt.Printf("  StatusInactive = %d (已停用)\n", StatusInactive)
	fmt.Printf("  StatusDeleted  = %d (已删除)\n", StatusDeleted)

	fmt.Println("权限位标志:")
	perm := Read | Write // 权限组合: 可读可写
	fmt.Printf("  Read | Write = %d (可读可写)\n", perm)
	fmt.Printf("  可执行? %t\n", perm&Execute != 0) // 没有 Execute 权限

	// ============================================================
	// 6. 类型转换（Type Conversion）— Go 没有隐式类型转换
	// ============================================================
	fmt.Println("\n=== 类型转换 ===")
	var intVal int = 42
	var floatVal float64 = float64(intVal) // int → float64
	fmt.Printf("int=%d → float64=%.1f\n", intVal, floatVal)

	// float64 → int（截断小数）
	var piFloat float64 = 3.1415926
	var piInt int = int(piFloat)
	fmt.Printf("float64=%.4f → int=%d (截断小数)\n", piFloat, piInt)

	// int → string（注意：这是 Unicode 码点，不是数字字符串）
	var codePoint int = 65 // 'A' 的 ASCII
	char := string(rune(codePoint))
	fmt.Printf("int=%d → string(rune)=%q\n", codePoint, char)

	// 正确的数字→字符串转换需要使用 strconv, 但这里用简单方式
	// 使用 fmt.Sprintf
	strFromInt := fmt.Sprintf("%d", 12345)
	fmt.Printf("int=12345 → string=%q (通过 Sprintf)\n", strFromInt)

	// 字符串 ↔ rune slice
	hello := "你好"
	runes := []rune(hello)
	fmt.Printf("string=%q → []rune=%v\n", hello, runes)

	// 数值精度损失示例
	large := int64(math.MaxInt64)
	converted := int64(float64(large)) // float64 精度只有 53 bit
	fmt.Printf("int64 精度损失: %d → %d (差=%d)\n", large, converted, large-converted)

	// ============================================================
	// 7. 命名规范 — 演示导出 vs 未导出
	// ============================================================
	fmt.Println("\n=== 命名规范 ===")

	// 局部变量使用驼峰 (camelCase)
	userName := "王五"
	userAge := 28
	fmt.Printf("局部变量 (camelCase): userName=%s, userAge=%d\n", userName, userAge)

	// 包级别未导出（小写开头）— 仅包内可用
	fmt.Printf("包内常量: %s\n", unexportedConst)

	// 导出标识符（大写开头）— 其他包可引用
	fmt.Printf("导出常量 Pi=%.4f\n", Pi)

	// 缩写词全大写（例如 HTTP, URL, ID）
	userID := "u-1001"
	httpStatusCode := 200
	budgetURL := "https://example.com/budget"
	fmt.Printf("缩写惯例: userID=%s, httpStatusCode=%d, budgetURL=%s\n",
		userID, httpStatusCode, budgetURL)

	// 首字母缩略词一致性测试
	checkNaming("userID", isCamelCase("userID"))
	checkNaming("UserID", isCamelCase("UserID"))
}

// 辅助：除法，返回商和余数
func divide(a, b int) (int, int) {
	return a / b, a % b
}

// ============================================================
// 测试辅助函数 — 供 main_test.go 调用
// ============================================================

// GetZeroValues 返回所有零值的字符串表示，用于测试零值行为
func GetZeroValues() map[string]string {
	return map[string]string{
		"int":     fmt.Sprintf("%d", zeroInt),
		"float64": fmt.Sprintf("%f", zeroFloat),
		"bool":    fmt.Sprintf("%t", zeroBool),
		"string":  fmt.Sprintf("%q", zeroString),
	}
}

// DemonstrateConversions 执行类型转换并返回关键结果
func DemonstrateConversions() map[string]string {
	result := make(map[string]string)
	var intVal int = 42
	result["int_to_float"] = fmt.Sprintf("%.1f", float64(intVal))

	var piFloat float64 = 3.1415926
	result["float_to_int"] = fmt.Sprintf("%d", int(piFloat))

	result["rune_to_string"] = fmt.Sprintf("%q", string(rune(65)))
	return result
}

// GetIotaConstants 返回 iota 枚举常量值，用于验证 iota 行为
func GetIotaConstants() map[string]int {
	return map[string]int{
		"StatusPending":  StatusPending,
		"StatusActive":   StatusActive,
		"StatusInactive": StatusInactive,
		"StatusDeleted":  StatusDeleted,
	}
}

// GetPermissionFlags 返回权限位标志值，用于验证位运算
func GetPermissionFlags() map[string]int {
	return map[string]int{
		"Read":    Read,
		"Write":   Write,
		"Execute": Execute,
	}
}

// SwapValues 交换两个整数的值（通过指针），用于测试多重赋值
func SwapValues(a, b int) (int, int) {
	return b, a
}

// Divide 公开的除法函数，用于测试多返回值
func Divide(a, b int) (int, int) {
	return divide(a, b)
}

// IsCamelCase 公开版 isCamelCase，用于测试
func IsCamelCase(s string) bool {
	return isCamelCase(s)
}

// 辅助：检查命名是否符合驼峰（仅演示用）
func isCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// 简单检查：首字母是小写就是驼峰
	return unicode.IsLower(rune(s[0]))
}

func checkNaming(name string, ok bool) {
	if ok {
		fmt.Printf("  %s ✅ 驼峰命名 (未导出)\n", name)
	} else {
		fmt.Printf("  %s 🔺 首字母大写 (导出/包外可见)\n", name)
	}
}