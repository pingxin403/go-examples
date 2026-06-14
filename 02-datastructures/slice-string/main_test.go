// main_test.go — 对 slice-string 包的核心操作进行表格驱动测试
//
// 覆盖范围：
//   - 切片创建、切片表达式、append/copy
//   - 删除、插入、过滤、反转
//   - 字符串 rune/byte 操作、拼接方式
package main

import (
	"testing"
)

// ============================================================
// 1. 切片创建与属性
// ============================================================

// TestMakeSlice 测试 MakeSlice 创建切片的长度和容量
func TestMakeSlice(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		capacity int
		vals     []int
		wantLen  int
		wantCap  int
		want0    int // s[0] 的值
	}{
		{name: "基本创建", length: 3, capacity: 5, vals: []int{10, 20, 30}, wantLen: 3, wantCap: 5, want0: 10},
		{name: "零长切片", length: 0, capacity: 0, vals: nil, wantLen: 0, wantCap: 0, want0: 0},
		{name: "部分填充", length: 5, capacity: 10, vals: []int{1, 2}, wantLen: 5, wantCap: 10, want0: 1},
		{name: "值多于长度", length: 2, capacity: 5, vals: []int{100, 200, 300}, wantLen: 2, wantCap: 5, want0: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeSlice(tt.length, tt.capacity, tt.vals...)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, 期望 %d", len(got), tt.wantLen)
			}
			if cap(got) != tt.wantCap {
				t.Errorf("cap = %d, 期望 %d", cap(got), tt.wantCap)
			}
			if tt.want0 != 0 && len(got) > 0 && got[0] != tt.want0 {
				t.Errorf("got[0] = %d, 期望 %d", got[0], tt.want0)
			}
		})
	}
}

// TestSliceLenCap 测试 SliceLenCap 返回正确的长度和容量
func TestSliceLenCap(t *testing.T) {
	tests := []struct {
		name    string
		s       []int
		wantLen int
		wantCap int
	}{
		{name: "正常切片", s: []int{1, 2, 3}, wantLen: 3, wantCap: 3},
		{name: "空切片", s: []int{}, wantLen: 0, wantCap: 0},
		{name: "nil 切片", s: nil, wantLen: 0, wantCap: 0},
		{name: "大切片", s: make([]int, 10, 20), wantLen: 10, wantCap: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLen, gotCap := SliceLenCap(tt.s)
			if gotLen != tt.wantLen {
				t.Errorf("len = %d, 期望 %d", gotLen, tt.wantLen)
			}
			if gotCap != tt.wantCap {
				t.Errorf("cap = %d, 期望 %d", gotCap, tt.wantCap)
			}
		})
	}
}

// ============================================================
// 2. 切片表达式
// ============================================================

// TestSubSlice 测试切片表达式操作
func TestSubSlice(t *testing.T) {
	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	tests := []struct {
		name      string
		start, end int
		wantLen   int
		wantFirst int // 子切片第一个元素
	}{
		{name: "中间范围 2:5", start: 2, end: 5, wantLen: 3, wantFirst: 2},
		{name: "从头到4", start: 0, end: 4, wantLen: 4, wantFirst: 0},
		{name: "从6到尾", start: 6, end: 10, wantLen: 4, wantFirst: 6},
		{name: "全部", start: 0, end: 10, wantLen: 10, wantFirst: 0},
		{name: "单元素", start: 3, end: 4, wantLen: 1, wantFirst: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubSlice(s, tt.start, tt.end)
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, 期望 %d", len(got), tt.wantLen)
			}
			if len(got) > 0 && got[0] != tt.wantFirst {
				t.Errorf("got[0] = %d, 期望 %d", got[0], tt.wantFirst)
			}
		})
	}
}

// TestSubSlice_边界情况 测试切片表达式的边界与非法输入
func TestSubSlice_边界情况(t *testing.T) {
	s := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	t.Run("负数起始返回 nil", func(t *testing.T) {
		got := SubSlice(s, -1, 5)
		if got != nil {
			t.Error("期望 nil 切片")
		}
	})

	t.Run("结束位置越界返回 nil", func(t *testing.T) {
		got := SubSlice(s, 0, 20)
		if got != nil {
			t.Error("期望 nil 切片")
		}
	})

	t.Run("起始大于结束返回 nil", func(t *testing.T) {
		got := SubSlice(s, 5, 2)
		if got != nil {
			t.Error("期望 nil 切片")
		}
	})
}

// ============================================================
// 3. Append 与 Copy
// ============================================================

// TestAppendValues 测试 append 操作
func TestAppendValues(t *testing.T) {
	t.Run("追加多个值", func(t *testing.T) {
		s := []int{1, 2, 3}
		result, l, _ := AppendValues(s, 4, 5, 6)
		if l != 6 {
			t.Errorf("len = %d, 期望 6", l)
		}
		if result[3] != 4 || result[4] != 5 || result[5] != 6 {
			t.Errorf("追加结果不正确: %v", result)
		}
	})

	t.Run("追加到 nil 切片", func(t *testing.T) {
		var s []int
		result, l, _ := AppendValues(s, 1, 2)
		if l != 2 {
			t.Errorf("len = %d, 期望 2", l)
		}
		if result[0] != 1 || result[1] != 2 {
			t.Errorf("追加结果不正确: %v", result)
		}
	})

	t.Run("不追加值", func(t *testing.T) {
		s := []int{1, 2}
		result, l, _ := AppendValues(s)
		if l != 2 {
			t.Errorf("len = %d, 期望 2", l)
		}
		if result[0] != 1 {
			t.Errorf("result[0] = %d, 期望 1", result[0])
		}
	})
}

// TestCopySlice 测试 copy 操作
func TestCopySlice(t *testing.T) {
	t.Run("完整复制", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := make([]int, 3)
		n := CopySlice(b, a)
		if n != 3 {
			t.Errorf("复制个数 = %d, 期望 3", n)
		}
		if b[0] != 1 || b[1] != 2 || b[2] != 3 {
			t.Errorf("复制结果不正确: %v", b)
		}
	})

	t.Run("目标更小", func(t *testing.T) {
		a := []int{1, 2, 3}
		small := make([]int, 2)
		n := CopySlice(small, a)
		if n != 2 {
			t.Errorf("复制个数 = %d, 期望 2", n)
		}
		if small[0] != 1 || small[1] != 2 {
			t.Errorf("复制结果不正确: %v", small)
		}
	})

	t.Run("源为 nil", func(t *testing.T) {
		dst := make([]int, 3)
		n := CopySlice(dst, nil)
		if n != 0 {
			t.Errorf("复制个数 = %d, 期望 0", n)
		}
	})
}

// ============================================================
// 4. 删除元素
// ============================================================

// TestRemoveIndex 测试保持顺序的删除操作
func TestRemoveIndex(t *testing.T) {
	tests := []struct {
		name    string
		s       []int
		idx     int
		want    []int
		wantOK  bool
	}{
		{name: "删除中间元素", s: []int{1, 2, 3, 4, 5}, idx: 2, want: []int{1, 2, 4, 5}, wantOK: true},
		{name: "删除第一个", s: []int{1, 2, 3}, idx: 0, want: []int{2, 3}, wantOK: true},
		{name: "删除最后一个", s: []int{1, 2, 3}, idx: 2, want: []int{1, 2}, wantOK: true},
		{name: "索引越界（负数）", s: []int{1, 2, 3}, idx: -1, want: []int{1, 2, 3}, wantOK: false},
		{name: "索引越界（超出）", s: []int{1, 2, 3}, idx: 5, want: []int{1, 2, 3}, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := RemoveIndex(tt.s, tt.idx)
			if ok != tt.wantOK {
				t.Errorf("ok = %t, 期望 %t", ok, tt.wantOK)
			}
			if !sliceEqual(got, tt.want) {
				t.Errorf("got = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

// TestRemoveIndexFast 测试不保持顺序的快速删除
func TestRemoveIndexFast(t *testing.T) {
	tests := []struct {
		name   string
		s      []int
		idx    int
		wantOK bool
	}{
		{name: "删除中间元素", s: []int{1, 2, 3, 4, 5}, idx: 2, wantOK: true},
		{name: "删除唯一元素", s: []int{42}, idx: 0, wantOK: true},
		{name: "索引越界", s: []int{1, 2, 3}, idx: 10, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := RemoveIndexFast(tt.s, tt.idx)
			if ok != tt.wantOK {
				t.Errorf("ok = %t, 期望 %t", ok, tt.wantOK)
			}
			if ok && len(got) != len(tt.s)-1 {
				t.Errorf("len = %d, 期望 %d", len(got), len(tt.s)-1)
			}
		})
	}
}

// ============================================================
// 5. 插入与过滤
// ============================================================

// TestInsertValue 测试在指定位置插入元素
func TestInsertValue(t *testing.T) {
	tests := []struct {
		name   string
		s      []int
		idx    int
		val    int
		want   []int
		wantOK bool
	}{
		{name: "插入到中间", s: []int{1, 2, 4, 5}, idx: 2, val: 3, want: []int{1, 2, 3, 4, 5}, wantOK: true},
		{name: "插入到开头", s: []int{2, 3, 4}, idx: 0, val: 1, want: []int{1, 2, 3, 4}, wantOK: true},
		{name: "插入到末尾", s: []int{1, 2, 3}, idx: 3, val: 4, want: []int{1, 2, 3, 4}, wantOK: true},
		{name: "索引越界（负数）", s: []int{1, 2}, idx: -1, val: 99, want: []int{1, 2}, wantOK: false},
		{name: "索引越界（超出）", s: []int{1, 2}, idx: 5, val: 99, want: []int{1, 2}, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := InsertValue(tt.s, tt.idx, tt.val)
			if ok != tt.wantOK {
				t.Errorf("ok = %t, 期望 %t", ok, tt.wantOK)
			}
			if !sliceEqual(got, tt.want) {
				t.Errorf("got = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

// TestFilterEven 测试原地过滤偶数
func TestFilterEven(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		want []int
	}{
		{name: "混合奇偶", s: []int{1, 2, 3, 4, 5, 6}, want: []int{2, 4, 6}},
		{name: "全部偶数", s: []int{2, 4, 6}, want: []int{2, 4, 6}},
		{name: "全部奇数", s: []int{1, 3, 5}, want: []int{}},
		{name: "空切片", s: []int{}, want: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterEven(tt.s)
			if !sliceEqual(got, tt.want) {
				t.Errorf("got = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

// TestReverseSlice 测试切片反转
func TestReverseSlice(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		want []int
	}{
		{name: "奇数个元素", s: []int{1, 2, 3, 4, 5}, want: []int{5, 4, 3, 2, 1}},
		{name: "偶数个元素", s: []int{1, 2, 3, 4}, want: []int{4, 3, 2, 1}},
		{name: "单元素", s: []int{42}, want: []int{42}},
		{name: "空切片", s: []int{}, want: []int{}},
		{name: "两个元素", s: []int{1, 2}, want: []int{2, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReverseSlice(tt.s)
			if !sliceEqual(tt.s, tt.want) {
				t.Errorf("got = %v, 期望 %v", tt.s, tt.want)
			}
		})
	}
}

// ============================================================
// 6. 字符串操作
// ============================================================

// TestRuneCount 测试字符串的字符数计算
func TestRuneCount(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{name: "ASCII 字符串", s: "Hello", want: 5},
		{name: "中文字符串", s: "世界", want: 2},
		{name: "混合字符串", s: "Hello, 世界", want: 9},
		{name: "空字符串", s: "", want: 0},
		{name: "表情符号", s: "😀🚀", want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuneCount(tt.s)
			if got != tt.want {
				t.Errorf("RuneCount(%q) = %d, 期望 %d", tt.s, got, tt.want)
			}
		})
	}
}

// TestByteLen 测试字符串的字节数（注意中文多字节）
func TestByteLen(t *testing.T) {
	t.Run("ASCII 字符串字节数等于字符数", func(t *testing.T) {
		got := ByteLen("Hello")
		if got != 5 {
			t.Errorf("ByteLen = %d, 期望 5", got)
		}
	})

	t.Run("中文每个字符占 3 字节", func(t *testing.T) {
		got := ByteLen("世界")
		if got != 6 {
			t.Errorf("ByteLen = %d, 期望 6", got)
		}
	})

	t.Run("空字符串字节数为 0", func(t *testing.T) {
		got := ByteLen("")
		if got != 0 {
			t.Errorf("ByteLen = %d, 期望 0", got)
		}
	})
}

// TestConcatWithParts 测试使用 + 拼接字符串
func TestConcatWithParts(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{name: "拼接三个部分", parts: []string{"Hello", "World"}, want: "Hello, World"},
		{name: "单一部分", parts: []string{"Go"}, want: "Go"},
		{name: "空输入", parts: []string{}, want: ""},
		{name: "多个部分", parts: []string{"a", "b", "c"}, want: "a, b, c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConcatWithParts(tt.parts...)
			if got != tt.want {
				t.Errorf("ConcatWithParts(%v) = %q, 期望 %q", tt.parts, got, tt.want)
			}
		})
	}
}

// TestConcatWithBuilder 测试使用 strings.Builder 拼接字符串
func TestConcatWithBuilder(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{name: "拼接三个部分", parts: []string{"Hello", "World"}, want: "Hello, World"},
		{name: "单一部分", parts: []string{"Go"}, want: "Go"},
		{name: "空输入", parts: []string{}, want: ""},
		{name: "多个部分", parts: []string{"a", "b", "c"}, want: "a, b, c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConcatWithBuilder(tt.parts...)
			if got != tt.want {
				t.Errorf("ConcatWithBuilder(%v) = %q, 期望 %q", tt.parts, got, tt.want)
			}
		})
	}
}

// TestConcat_两种方式结果一致 验证两种拼接方式输出相同
func TestConcat_两种方式结果一致(t *testing.T) {
	parts := []string{"Go", "Java", "Python", "Rust"}
	a := ConcatWithParts(parts...)
	b := ConcatWithBuilder(parts...)
	if a != b {
		t.Errorf("两种拼接方式结果不一致: +=%q, Builder=%q", a, b)
	}
}

// TestToBytes 测试字符串转 []byte
func TestToBytes(t *testing.T) {
	s := "Hello"
	b := ToBytes(s)
	if len(b) != 5 {
		t.Errorf("len = %d, 期望 5", len(b))
	}
	if b[0] != 'H' {
		t.Errorf("b[0] = %d, 期望 %d", b[0], 'H')
	}

	// 验证中文 UTF-8 编码
	s2 := "世"
	b2 := ToBytes(s2)
	if len(b2) != 3 {
		t.Errorf("'世' 的 UTF-8 字节数 = %d, 期望 3", len(b2))
	}
}

// TestToRunes 测试字符串转 []rune
func TestToRunes(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int // rune 个数
	}{
		{name: "ASCII", s: "Hi", want: 2},
		{name: "中文", s: "世界你好", want: 4},
		{name: "混合", s: "Go语言", want: 4},
		{name: "空字符串", s: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRunes(tt.s)
			if len(got) != tt.want {
				t.Errorf("len(runes) = %d, 期望 %d, 输入=%q", len(got), tt.want, tt.s)
			}
		})
	}
}

// ============================================================
// 辅助函数
// ============================================================

// sliceEqual 比较两个 int 切片是否相等
func sliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}