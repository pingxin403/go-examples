package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ============================================================
// 可测试的 Helper 函数（供测试文件调用）
// ============================================================

// MakeSlice 使用 make 创建指定长度和容量的 int 切片并填充值
func MakeSlice(length, capacity int, vals ...int) []int {
	s := make([]int, length, capacity)
	for i := 0; i < len(s) && i < len(vals); i++ {
		s[i] = vals[i]
	}
	return s
}

// SliceLenCap 返回切片的长度和容量
func SliceLenCap(s []int) (int, int) {
	return len(s), cap(s)
}

// SubSlice 对切片执行切片表达式，返回指定范围的子切片
func SubSlice(s []int, start, end int) []int {
	if start < 0 || end > len(s) || start > end {
		return nil
	}
	return s[start:end]
}

// AppendValues 向切片追加多个值，返回新切片以及追加后的 len 和 cap
func AppendValues(s []int, vals ...int) ([]int, int, int) {
	result := append(s, vals...)
	return result, len(result), cap(result)
}

// CopySlice 复制 src 到目标切片并返回复制元素个数
func CopySlice(dst, src []int) int {
	return copy(dst, src)
}

// RemoveIndex 删除切片中指定索引的元素（保持顺序）
func RemoveIndex(s []int, idx int) ([]int, bool) {
	if idx < 0 || idx >= len(s) {
		return s, false
	}
	return append(s[:idx], s[idx+1:]...), true
}

// RemoveIndexFast 删除切片中指定索引的元素（不保持顺序，速度快）
func RemoveIndexFast(s []int, idx int) ([]int, bool) {
	if idx < 0 || idx >= len(s) {
		return s, false
	}
	s[idx] = s[len(s)-1]
	return s[:len(s)-1], true
}

// InsertValue 在切片指定位置插入一个值
func InsertValue(s []int, idx, val int) ([]int, bool) {
	if idx < 0 || idx > len(s) {
		return s, false
	}
	result := append(s[:idx], append([]int{val}, s[idx:]...)...)
	return result, true
}

// FilterEven 过滤出切片中的偶数（原地过滤，返回子切片）
func FilterEven(s []int) []int {
	filtered := s[:0]
	for _, v := range s {
		if v%2 == 0 {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// ReverseSlice 原地反转切片
func ReverseSlice(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// RuneCount 返回字符串的字符（rune）数
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// ByteLen 返回字符串的字节数
func ByteLen(s string) int {
	return len(s)
}

// ConcatWithPlus 使用 + 操作符拼接字符串
func ConcatWithParts(parts ...string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// ConcatWithBuilder 使用 strings.Builder 高效拼接字符串
func ConcatWithBuilder(parts ...string) string {
	var sb strings.Builder
	for i, p := range parts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p)
	}
	return sb.String()
}

// ToBytes 将字符串转为 []byte
func ToBytes(s string) []byte {
	return []byte(s)
}

// ToRunes 将字符串转为 []rune
func ToRunes(s string) []rune {
	return []rune(s)
}

// ------------------------------------------------------------
// 1. 切片（Slice）基础操作
// ------------------------------------------------------------

// demoSliceCreate 演示切片的创建方式：make 和字面量
func demoSliceCreate() {
	fmt.Println("=== 切片创建 ===")

	// make 创建：长度 3，容量 5
	s1 := make([]int, 3, 5)
	s1[0] = 10
	s1[1] = 20
	s1[2] = 30
	fmt.Printf("make:   %v (len=%d, cap=%d)\n", s1, len(s1), cap(s1))

	// 字面量创建
	s2 := []int{1, 2, 3, 4, 5}
	fmt.Printf("literal: %v (len=%d, cap=%d)\n", s2, len(s2), cap(s2))

	// nil 切片 vs 空切片
	var nilSlice []int
	emptySlice := []int{}
	fmt.Printf("nil slice:  %v (len=%d, cap=%d, isNil=%t)\n", nilSlice, len(nilSlice), cap(nilSlice), nilSlice == nil)
	fmt.Printf("empty slice: %v (len=%d, cap=%d, isNil=%t)\n", emptySlice, len(emptySlice), cap(emptySlice), emptySlice == nil)
}

// demoSlicingExpr 演示切片表达式（slicing expression）
func demoSlicingExpr() {
	fmt.Println("\n=== 切片表达式 ===")

	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	sub1 := s[2:5]           // [2,3,4]
	sub2 := s[:4]            // [0,1,2,3]
	sub3 := s[6:]            // [6,7,8,9]
	sub4 := s[:]             // 全部
	fmt.Printf("s[2:5] = %v (len=%d, cap=%d)\n", sub1, len(sub1), cap(sub1))
	fmt.Printf("s[:4]  = %v (len=%d, cap=%d)\n", sub2, len(sub2), cap(sub2))
	fmt.Printf("s[6:]  = %v (len=%d, cap=%d)\n", sub3, len(sub3), cap(sub3))
	fmt.Printf("s[:]   = %v\n", sub4)

	// 修改子切片会影响底层数组
	sub1[0] = 999
	fmt.Printf("修改子切片后原切片: %v\n", s)
}

// demoAppendCopy 演示 append 和 copy
func demoAppendCopy() {
	fmt.Println("\n=== append & copy ===")

	// append 自动扩容
	var s []int
	for i := 0; i < 10; i++ {
		s = append(s, i)
		fmt.Printf("append %d: len=%d cap=%d\n", i, len(s), cap(s))
	}

	// copy
	a := []int{1, 2, 3}
	b := make([]int, len(a))
	n := copy(b, a)
	fmt.Printf("copy: a=%v b=%v n=%d\n", a, b, n)

	// copy 只复制 min(len(dst), len(src)) 个元素
	small := make([]int, 2)
	copy(small, a)
	fmt.Printf("copy into smaller: %v\n", small)
}

// demoSliceTricks 演示常见切片技巧
func demoSliceTricks() {
	fmt.Println("\n=== 切片技巧 ===")

	// -- 删除（保持顺序） --
	s := []int{1, 2, 3, 4, 5}
	idx := 2
	s = append(s[:idx], s[idx+1:]...)
	fmt.Printf("删除 index=2: %v\n", s)

	// -- 删除（不保持顺序）--
	s2 := []int{1, 2, 3, 4, 5}
	s2[idx] = s2[len(s2)-1]
	s2 = s2[:len(s2)-1]
	fmt.Printf("删除(快删) index=2: %v\n", s2)

	// -- 插入 --
	s3 := []int{1, 2, 4, 5}
	insPos := 2
	s3 = append(s3[:insPos], append([]int{3}, s3[insPos:]...)...)
	fmt.Printf("插入 3 到 index=2: %v\n", s3)

	// -- 过滤 --
	nums := []int{1, 2, 3, 4, 5, 6}
	filtered := nums[:0]
	for _, v := range nums {
		if v%2 == 0 {
			filtered = append(filtered, v)
		}
	}
	fmt.Printf("过滤偶数: %v (原地过滤)\n", filtered)

	// -- 反转 --
	rev := []int{1, 2, 3, 4, 5}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	fmt.Printf("反转: %v\n", rev)
}

// ------------------------------------------------------------
// 2. 字符串（String）操作
// ------------------------------------------------------------

// demoStringBasics 演示字符串作为只读字节切片
func demoStringBasics() {
	fmt.Println("\n=== 字符串基础 ===")

	s := "Hello, 世界"
	fmt.Printf("字符串: %q\n", s)
	fmt.Printf("len(s) = %d (字节数, 不是字符数)\n", len(s))
	fmt.Printf("utf8.RuneCountInString(s) = %d (字符数)\n", utf8.RuneCountInString(s))

	// 按字节遍历
	fmt.Print("字节遍历: ")
	for i := 0; i < len(s); i++ {
		fmt.Printf("%02x ", s[i])
	}
	fmt.Println()

	// 按 rune 遍历
	fmt.Print("rune遍历: ")
	for i, r := range s {
		fmt.Printf("[%d:%c(%U)] ", i, r, r)
	}
	fmt.Println()
}

// demoRuneVsByte 演示 rune 和 byte 的区别
func demoRuneVsByte() {
	fmt.Println("\n=== rune vs byte ===")

	s := "世"

	// byte 层面（UTF-8 编码，3 个字节）
	bs := []byte(s)
	fmt.Printf("[]byte(%q) = %v (长度 %d)\n", s, bs, len(bs))

	// rune 层面（Unicode 码点，1 个 rune）
	rs := []rune(s)
	fmt.Printf("[]rune(%q) = %v (长度 %d, U+%04X)\n", s, rs, len(rs), rs[0])
}

// demoStringBuild 演示字符串拼接的几种方式
func demoStringBuild() {
	fmt.Println("\n=== 字符串拼接 ===")

	// 方式 1: + 操作符（小规模可用）
	s1 := "Hello" + ", " + "World"
	fmt.Printf("(+) 操作符: %q\n", s1)

	// 方式 2: strings.Builder（推荐，高性能）
	var sb strings.Builder
	sb.WriteString("Hello")
	sb.WriteString(", ")
	sb.WriteString("World")
	_ = sb.WriteByte('!')
	fmt.Printf("strings.Builder: %q (len=%d)\n", sb.String(), sb.Len())

	// 方式 3: []byte 转换
	b := []byte("Hello")
	b = append(b, ", "...)
	b = append(b, "World"...)
	fmt.Printf("[]byte: %q\n", string(b))

	// 性能对比说明
	fmt.Println("--- 性能建议 ---")
	fmt.Println("少量拼接: + 操作符简洁够用")
	fmt.Println("大量拼接: strings.Builder 最佳")
	fmt.Println("需要 []byte 操作: 直接使用 byte slice")
}

// ------------------------------------------------------------
// main
// ------------------------------------------------------------

func main() {
	fmt.Println("========================================")
	fmt.Println("  02-datastructures / slice-string")
	fmt.Println("========================================")

	demoSliceCreate()
	demoSlicingExpr()
	demoAppendCopy()
	demoSliceTricks()
	demoStringBasics()
	demoRuneVsByte()
	demoStringBuild()

	fmt.Println("\n=== 完成 ===")
}