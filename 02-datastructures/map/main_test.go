// main_test.go — 对 map 包的核心操作进行表格驱动测试
//
// 覆盖范围：
//   - Map 创建、CRUD（增删改查）、comma-ok 惯用法
//   - Map 迭代与 key 排序
//   - Map 作为 Set 使用
//   - Struct 作为 Map Key
//   - 词频统计实战
package main

import (
	"sort"
	"testing"
)

// ============================================================
// 1. Map 创建与 CRUD
// ============================================================

// TestMapPut 测试向 map 中写入键值对
func TestMapPut(t *testing.T) {
	t.Run("插入新 key", func(t *testing.T) {
		m := make(map[string]int)
		old, exists := MapPut(m, "apple", 5)
		if exists {
			t.Error("apple 不应已存在")
		}
		if old != 0 {
			t.Errorf("old = %d, 期望 0", old)
		}
		if m["apple"] != 5 {
			t.Errorf("m[apple] = %d, 期望 5", m["apple"])
		}
	})

	t.Run("覆盖已有 key", func(t *testing.T) {
		m := map[string]int{"apple": 5}
		old, exists := MapPut(m, "apple", 10)
		if !exists {
			t.Error("apple 应已存在")
		}
		if old != 5 {
			t.Errorf("old = %d, 期望 5", old)
		}
		if m["apple"] != 10 {
			t.Errorf("m[apple] = %d, 期望 10", m["apple"])
		}
	})
}

// TestMapGet 测试从 map 中读取值
func TestMapGet(t *testing.T) {
	t.Run("读取存在的 key", func(t *testing.T) {
		m := map[string]int{"apple": 10}
		v, ok := MapGet(m, "apple")
		if !ok {
			t.Error("apple 应存在")
		}
		if v != 10 {
			t.Errorf("v = %d, 期望 10", v)
		}
	})

	t.Run("读取不存在的 key", func(t *testing.T) {
		m := map[string]int{"apple": 10}
		v, ok := MapGet(m, "banana")
		if ok {
			t.Error("banana 不应存在")
		}
		if v != 0 {
			t.Errorf("v = %d, 期望 0（零值）", v)
		}
	})

	t.Run("从空 map 读取", func(t *testing.T) {
		m := make(map[string]int)
		v, ok := MapGet(m, "anything")
		if ok {
			t.Error("anything 不应存在")
		}
		if v != 0 {
			t.Errorf("v = %d, 期望 0", v)
		}
	})
}

// TestMapDelete 测试从 map 中删除 key
func TestMapDelete(t *testing.T) {
	t.Run("删除存在的 key", func(t *testing.T) {
		m := map[string]int{"apple": 5, "banana": 3}
		existed := MapDelete(m, "apple")
		if !existed {
			t.Error("apple 应存在")
		}
		if _, ok := m["apple"]; ok {
			t.Error("apple 应已被删除")
		}
		if len(m) != 1 {
			t.Errorf("len = %d, 期望 1", len(m))
		}
	})

	t.Run("删除不存在的 key", func(t *testing.T) {
		m := map[string]int{"apple": 5}
		existed := MapDelete(m, "banana")
		if existed {
			t.Error("banana 不应存在")
		}
	})

	t.Run("从空 map 删除", func(t *testing.T) {
		m := make(map[string]int)
		existed := MapDelete(m, "anything")
		if existed {
			t.Error("anything 不应存在")
		}
	})
}

// ============================================================
// 2. Map 迭代与 Key 排序
// ============================================================

// TestMapKeys 测试返回排序后的 keys
func TestMapKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]int
		want []string
	}{
		{name: "多个 key", m: map[string]int{"b": 2, "a": 1, "c": 3}, want: []string{"a", "b", "c"}},
		{name: "单个 key", m: map[string]int{"x": 10}, want: []string{"x"}},
		{name: "空 map", m: map[string]int{}, want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapKeys(tt.m)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, 期望 %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %s, 期望 %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestMapKeys_排序保证 验证返回结果确实是排序的
func TestMapKeys_排序保证(t *testing.T) {
	m := map[string]int{
		"orange": 3,
		"apple":  5,
		"banana": 2,
		"cherry": 7,
		"date":   1,
	}
	keys := MapKeys(m)
	if !sort.StringsAreSorted(keys) {
		t.Errorf("keys 未排序: %v", keys)
	}
	// 验证每个 key 在原 map 中都能找到
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Errorf("key %q 不在原 map 中", k)
		}
	}
}

// ============================================================
// 3. Map 作为 Set
// ============================================================

// TestNewStringSet 测试创建字符串集合
func TestNewStringSet(t *testing.T) {
	t.Run("从多个元素创建", func(t *testing.T) {
		set := NewStringSet("apple", "banana", "orange")
		if len(set) != 3 {
			t.Errorf("len = %d, 期望 3", len(set))
		}
	})

	t.Run("从空参数创建", func(t *testing.T) {
		set := NewStringSet()
		if len(set) != 0 {
			t.Errorf("len = %d, 期望 0", len(set))
		}
	})

	t.Run("自动去重", func(t *testing.T) {
		set := NewStringSet("a", "a", "b")
		if len(set) != 2 {
			t.Errorf("有重复时 len = %d, 期望 2", len(set))
		}
	})
}

// TestSetAdd 测试向集合添加元素
func TestSetAdd(t *testing.T) {
	t.Run("添加新元素", func(t *testing.T) {
		set := NewStringSet("apple")
		existed := SetAdd(set, "banana")
		if existed {
			t.Error("banana 应是新元素")
		}
		if len(set) != 2 {
			t.Errorf("len = %d, 期望 2", len(set))
		}
	})

	t.Run("添加重复元素", func(t *testing.T) {
		set := NewStringSet("apple")
		existed := SetAdd(set, "apple")
		if !existed {
			t.Error("apple 应已存在")
		}
		if len(set) != 1 {
			t.Errorf("添加重复元素后 len = %d, 期望 1", len(set))
		}
	})
}

// TestSetHas 测试集合成员检查
func TestSetHas(t *testing.T) {
	t.Run("元素存在", func(t *testing.T) {
		set := NewStringSet("apple", "banana")
		if !SetHas(set, "apple") {
			t.Error("apple 应在集合中")
		}
	})

	t.Run("元素不存在", func(t *testing.T) {
		set := NewStringSet("apple", "banana")
		if SetHas(set, "grape") {
			t.Error("grape 不应在集合中")
		}
	})

	t.Run("空集合检查", func(t *testing.T) {
		set := NewStringSet()
		if SetHas(set, "anything") {
			t.Error("空集合中不应有元素")
		}
	})
}

// TestSetRemove 测试从集合中移除元素
func TestSetRemove(t *testing.T) {
	t.Run("移除存在的元素", func(t *testing.T) {
		set := NewStringSet("apple", "banana")
		existed := SetRemove(set, "apple")
		if !existed {
			t.Error("apple 应存在")
		}
		if SetHas(set, "apple") {
			t.Error("apple 应已被移除")
		}
		if len(set) != 1 {
			t.Errorf("len = %d, 期望 1", len(set))
		}
	})

	t.Run("移除不存在的元素", func(t *testing.T) {
		set := NewStringSet("apple")
		existed := SetRemove(set, "grape")
		if existed {
			t.Error("grape 不应存在")
		}
	})
}

// TestSetItems 测试获取集合中所有元素
func TestSetItems(t *testing.T) {
	set := NewStringSet("a", "b", "c")
	items := SetItems(set)
	if len(items) != 3 {
		t.Errorf("len = %d, 期望 3", len(items))
	}
	// 验证所有元素都在集合中
	for _, item := range items {
		if !SetHas(set, item) {
			t.Errorf("%q 应在集合中", item)
		}
	}

	t.Run("空集合", func(t *testing.T) {
		items := SetItems(NewStringSet())
		if len(items) != 0 {
			t.Errorf("len = %d, 期望 0", len(items))
		}
	})
}

// ============================================================
// 4. Struct 作为 Map Key
// ============================================================

// TestStructKey 测试结构体作为 map key
func TestStructKey(t *testing.T) {
	distances := map[Point]float64{
		{X: 0, Y: 0}: 0.0,
		{X: 3, Y: 4}: 5.0,
	}

	t.Run("查找存在的 key", func(t *testing.T) {
		d, ok := distances[Point{X: 3, Y: 4}]
		if !ok {
			t.Error("Point{3,4} 应在 map 中")
		}
		if d != 5.0 {
			t.Errorf("d = %.1f, 期望 5.0", d)
		}
	})

	t.Run("查找不存在的 key", func(t *testing.T) {
		_, ok := distances[Point{X: 1, Y: 1}]
		if ok {
			t.Error("Point{1,1} 不应在 map 中")
		}
	})

	t.Run("添加新 key", func(t *testing.T) {
		distances[Point{X: 5, Y: 12}] = 13.0
		d, ok := distances[Point{X: 5, Y: 12}]
		if !ok {
			t.Error("新添加的 key 应在 map 中")
		}
		if d != 13.0 {
			t.Errorf("d = %.1f, 期望 13.0", d)
		}
	})
}

// ============================================================
// 5. 词频统计
// ============================================================

// TestWordTokenize 测试文本分词
func TestWordTokenize(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int // 期望的单词数
	}{
		{name: "简单句子", text: "The quick brown fox", want: 4},
		{name: "带标点的句子", text: "Hello, world! How are you?", want: 5},
		{name: "空文本", text: "", want: 0},
		{name: "纯标点", text: "!!! ??? ,,,", want: 0},
		{name: "带数字", text: "Go 1 point 22 is released", want: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordTokenize(tt.text)
			if len(got) != tt.want {
				t.Errorf("单词数 = %d, 期望 %d; got=%v", len(got), tt.want, got)
			}
		})
	}
}

// TestWordFreq 测试词频统计
func TestWordFreq(t *testing.T) {
	t.Run("基本统计", func(t *testing.T) {
		text := "the cat and the dog"
		freq := WordFreq(text)
		if freq["the"] != 2 {
			t.Errorf("the 频率 = %d, 期望 2", freq["the"])
		}
		if freq["cat"] != 1 {
			t.Errorf("cat 频率 = %d, 期望 1", freq["cat"])
		}
		if freq["dog"] != 1 {
			t.Errorf("dog 频率 = %d, 期望 1", freq["dog"])
		}
	})

	t.Run("大小写不敏感", func(t *testing.T) {
		text := "The THE the"
		freq := WordFreq(text)
		if freq["the"] != 3 {
			t.Errorf("the 频率 = %d, 期望 3（大小写应归一化）", freq["the"])
		}
	})

	t.Run("空文本", func(t *testing.T) {
		freq := WordFreq("")
		if len(freq) != 0 {
			t.Errorf("空文本词频 map 长度 = %d, 期望 0", len(freq))
		}
	})

	t.Run("忽略标点", func(t *testing.T) {
		text := "hello, hello! hello?"
		freq := WordFreq(text)
		if freq["hello"] != 3 {
			t.Errorf("hello 频率 = %d, 期望 3", freq["hello"])
		}
	})
}

// ============================================================
// 6. 综合场景
// ============================================================

// TestMap_Set_综合场景 验证 map 和 set 配合使用
func TestMap_Set_综合场景(t *testing.T) {
	// 模拟一个简单的权限系统
	permissions := NewStringSet("read", "write", "delete")

	t.Run("检查权限", func(t *testing.T) {
		cases := []struct {
			perm string
			has  bool
		}{
			{"read", true},
			{"write", true},
			{"execute", false},
		}
		for _, c := range cases {
			got := SetHas(permissions, c.perm)
			if got != c.has {
				t.Errorf("权限 %q 存在=%t, 期望 %t", c.perm, got, c.has)
			}
		}
	})

	t.Run("动态添加权限", func(t *testing.T) {
		SetAdd(permissions, "execute")
		if !SetHas(permissions, "execute") {
			t.Error("execute 应已被添加")
		}
	})

	t.Run("动态移除权限", func(t *testing.T) {
		SetRemove(permissions, "delete")
		if SetHas(permissions, "delete") {
			t.Error("delete 应已被移除")
		}
	})
}