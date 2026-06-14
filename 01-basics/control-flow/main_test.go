// main_test.go — 对 control-flow 包中 if/for/switch/defer 等控制流特性的完整测试套件
//
// 本文件演示 Go 测试的多种模式：
//   - 表格驱动测试（Table-Driven Tests）
//   - 子测试（t.Run）
//   - 边界值测试
//   - 集合操作验证测试
package main

import (
	"sort"
	"testing"
	"time"
)

// ============================================================
// 1. if-else 逻辑测试
// ============================================================

// TestClassifyScore 表格驱动测试分数等级判断逻辑
func TestClassifyScore(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		{name: "90 分以上为优秀", score: 95, want: "优秀"},
		{name: "90 分为优秀", score: 90, want: "优秀"},
		{name: "80-89 为良好", score: 85, want: "良好"},
		{name: "80 分为良好", score: 80, want: "良好"},
		{name: "60-79 为及格", score: 60, want: "及格"},
		{name: "60 分为及格", score: 60, want: "及格"},
		{name: "59 分为不及格", score: 59, want: "不及格"},
		{name: "0 分为不及格", score: 0, want: "不及格"},
		{name: "负分处理", score: -5, want: "不及格"},
		{name: "满分 100", score: 100, want: "优秀"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyScore(tt.score)
			if got != tt.want {
				t.Errorf("ClassifyScore(%d) = %q; 期望 %q", tt.score, got, tt.want)
			}
		})
	}
}

// ============================================================
// 2. switch 表达式测试
// ============================================================

// TestIsWeekend 表格驱动测试周末判断
func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name string
		day  time.Weekday
		want bool
	}{
		{name: "周六是周末", day: time.Saturday, want: true},
		{name: "周日是周末", day: time.Sunday, want: true},
		{name: "周一是工作日", day: time.Monday, want: false},
		{name: "周二是工作日", day: time.Tuesday, want: false},
		{name: "周三是工作日", day: time.Wednesday, want: false},
		{name: "周四是工作日", day: time.Thursday, want: false},
		{name: "周五是工作日", day: time.Friday, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWeekend(tt.day)
			if got != tt.want {
				t.Errorf("IsWeekend(%v) = %t; 期望 %t", tt.day, got, tt.want)
			}
		})
	}
}

// ============================================================
// 3. for 循环测试
// ============================================================

// TestSumRange 表格驱动测试经典 for 循环求和
func TestSumRange(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{name: "1+2+3+4+5 = 15", n: 5, want: 15},
		{name: "1 到 1 求和 = 1", n: 1, want: 1},
		{name: "0 求和 = 0", n: 0, want: 0},
		{name: "1 到 10 求和 = 55", n: 10, want: 55},
		{name: "1 到 100 求和 = 5050", n: 100, want: 5050},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumRange(tt.n)
			if got != tt.want {
				t.Errorf("SumRange(%d) = %d; 期望 %d", tt.n, got, tt.want)
			}
		})
	}
}

// TestDoubleUntilLimit 表格驱动测试 while 风格循环
func TestDoubleUntilLimit(t *testing.T) {
	tests := []struct {
		name      string
		n, limit  int
		want      int
	}{
		{name: "1 加倍直到超过 100", n: 1, limit: 100, want: 128},
		{name: "初始值已超过 limit", n: 200, limit: 100, want: 200},
		{name: "2 加倍直到超过 50", n: 2, limit: 50, want: 64},
		{name: "limit 为 1", n: 1, limit: 1, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DoubleUntilLimit(tt.n, tt.limit)
			if got != tt.want {
				t.Errorf("DoubleUntilLimit(%d, %d) = %d; 期望 %d",
					tt.n, tt.limit, got, tt.want)
			}
		})
	}
}

// TestLoopWithBreak 表格驱动测试 break 语句
func TestLoopWithBreak(t *testing.T) {
	tests := []struct {
		name     string
		maxCount int
		want     int
	}{
		{name: "break 在 3 次后", maxCount: 3, want: 4},
		{name: "break 在 1 次后", maxCount: 1, want: 2},
		{name: "break 在 0 次后", maxCount: 0, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoopWithBreak(tt.maxCount)
			// count 从 1 开始，当 count > maxCount 时 break
			// 所以循环执行了 maxCount+1 次
			if got != tt.want {
				t.Errorf("LoopWithBreak(%d) = %d; 期望 %d",
					tt.maxCount, got, tt.want)
			}
		})
	}
}

// ============================================================
// 4. continue 测试
// ============================================================

// TestFilterEvens 表格驱动测试 continue 跳过奇数
func TestFilterEvens(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []int
	}{
		{name: "1 到 5 的偶数", n: 5, want: []int{2, 4}},
		{name: "1 到 10 的偶数", n: 10, want: []int{2, 4, 6, 8, 10}},
		{name: "1 到 1 无偶数", n: 1, want: []int{}},
		{name: "1 到 2 的偶数", n: 2, want: []int{2}},
		{name: "0 无偶数", n: 0, want: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterEvens(tt.n)
			if !slicesEqualInt(got, tt.want) {
				t.Errorf("FilterEvens(%d) = %v; 期望 %v", tt.n, got, tt.want)
			}
		})
	}
}

// ============================================================
// 5. for range 测试
// ============================================================

// TestRangeSum 表格驱动测试 for range 求和
func TestRangeSum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{name: "空 slice 求和为 0", nums: []int{}, want: 0},
		{name: "nil slice 求和为 0", nums: nil, want: 0},
		{name: "1+2+3=6", nums: []int{1, 2, 3}, want: 6},
		{name: "单个元素", nums: []int{42}, want: 42},
		{name: "包含负数", nums: []int{-5, 10, -3}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RangeSum(tt.nums)
			if got != tt.want {
				t.Errorf("RangeSum(%v) = %d; 期望 %d", tt.nums, got, tt.want)
			}
		})
	}
}

// TestMapKeysRange 验证 map key 遍历
func TestMapKeysRange(t *testing.T) {
	m := map[string]int{"张三": 92, "李四": 78, "王五": 88}
	keys := MapKeysRange(m)

	// 验证所有 key 都在结果中（顺序不固定）
	expectedKeys := []string{"张三", "李四", "王五"}
	if len(keys) != len(expectedKeys) {
		t.Fatalf("MapKeysRange 返回 %d 个 key; 期望 %d", len(keys), len(expectedKeys))
	}

	sort.Strings(keys)
	sort.Strings(expectedKeys)
	for i := range keys {
		if keys[i] != expectedKeys[i] {
			t.Errorf("key[%d] = %q; 期望 %q", i, keys[i], expectedKeys[i])
		}
	}
}

// TestCountRunes 表格驱动测试字符串 rune 遍历
func TestCountRunes(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{name: "空字符串", s: "", want: 0},
		{name: "ASCII 字符", s: "hello", want: 5},
		{name: "中文字符", s: "你好世界", want: 4},
		{name: "混合字符", s: "Hello 世界", want: 8},
		{name: "带表情符号", s: "Go 👍", want: 4}, // G(1) + o(1) +空格(1) + 👍(1) = 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountRunes(tt.s)
			if got != tt.want {
				t.Errorf("CountRunes(%q) = %d; 期望 %d", tt.s, got, tt.want)
			}
		})
	}
}

// TestChannelReceive 验证 channel range 遍历
func TestChannelReceive(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
	}{
		{name: "三条消息", messages: []string{"msg1", "msg2", "msg3"}},
		{name: "单条消息", messages: []string{"hello"}},
		{name: "空消息列表", messages: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChannelReceive(tt.messages)
			if len(got) != len(tt.messages) {
				t.Errorf("ChannelReceive 返回 %d 条消息; 期望 %d",
					len(got), len(tt.messages))
			}
			for i := range got {
				if got[i] != tt.messages[i] {
					t.Errorf("消息[%d] = %q; 期望 %q", i, got[i], tt.messages[i])
				}
			}
		})
	}
}

// ============================================================
// 6. switch fallthrough 测试
// ============================================================

// TestSwitchFallthrough 验证 fallthrough 行为
func TestSwitchFallthrough(t *testing.T) {
	tests := []struct {
		name string
		num  int
		want []string
	}{
		{name: "case 1 无 fallthrough", num: 1, want: []string{"case 1"}},
		{name: "case 2 有 fallthrough 到 case 3", num: 2, want: []string{"case 2", "case 3"}},
		{name: "case 3 正常不穿透", num: 3, want: []string{"case 3"}},
		{name: "case 4 走 default", num: 4, want: []string{"default"}},
		{name: "case 0 走 default", num: 0, want: []string{"default"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SwitchFallthrough(tt.num)
			if !slicesEqual(got, tt.want) {
				t.Errorf("SwitchFallthrough(%d) = %v; 期望 %v", tt.num, got, tt.want)
			}
		})
	}
}

// ============================================================
// 7. goto 测试
// ============================================================

// TestGotoSuccess 验证 goto 跳转到成功标签
func TestGotoSuccess(t *testing.T) {
	result := GotoSuccess()
	if result != "成功" {
		t.Errorf("GotoSuccess() = %q; 期望 %q", result, "成功")
	}
}

// ============================================================
// 辅助函数
// ============================================================

// slicesEqual 比较两个 string slice 是否相等
func slicesEqual(a, b []string) bool {
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

// slicesEqualInt 比较两个 int slice 是否相等
func slicesEqualInt(a, b []int) bool {
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