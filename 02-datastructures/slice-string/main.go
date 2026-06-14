package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

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