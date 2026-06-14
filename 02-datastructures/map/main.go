package main

import (
	"fmt"
	"sort"
	"strings"
)

// ============================================================
// 可测试的 Helper 函数（供测试文件调用）
// ============================================================

// MapPut 向 map 中写入 key-value，返回旧值（如有）和是否已存在
func MapPut(m map[string]int, key string, val int) (oldVal int, exists bool) {
	oldVal, exists = m[key]
	m[key] = val
	return
}

// MapGet 从 map 中读取值，使用 comma-ok 惯用法
func MapGet(m map[string]int, key string) (int, bool) {
	v, ok := m[key]
	return v, ok
}

// MapDelete 从 map 中删除 key，返回删除前是否存在
func MapDelete(m map[string]int, key string) bool {
	_, ok := m[key]
	delete(m, key)
	return ok
}

// MapKeys 返回 map 中所有 key 的已排序切片
func MapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// NewStringSet 创建一个基于 map[string]struct{} 的字符串集合
func NewStringSet(items ...string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

// SetAdd 向集合中添加元素，返回添加前是否已存在
func SetAdd(set map[string]struct{}, item string) bool {
	_, exists := set[item]
	set[item] = struct{}{}
	return exists
}

// SetHas 检查集合中是否包含指定元素
func SetHas(set map[string]struct{}, item string) bool {
	_, ok := set[item]
	return ok
}

// SetRemove 从集合中移除元素，返回移除前是否存在
func SetRemove(set map[string]struct{}, item string) bool {
	_, ok := set[item]
	delete(set, item)
	return ok
}

// SetItems 返回集合中所有元素的切片
func SetItems(set map[string]struct{}) []string {
	items := make([]string, 0, len(set))
	for k := range set {
		items = append(items, k)
	}
	return items
}

// WordTokenize 将文本按非字母数字字符分割为单词切片
func WordTokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	})
}

// WordFreq 统计单词频率，返回 map[string]int
func WordFreq(text string) map[string]int {
	words := WordTokenize(strings.ToLower(text))
	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++
	}
	return freq
}

// ------------------------------------------------------------
// 1. Map 创建与 CRUD
// ------------------------------------------------------------

// demoMapCreate 演示 map 的两种创建方式
func demoMapCreate() {
	fmt.Println("=== Map 创建 ===")

	// make 创建（动态添加）
	scores := make(map[string]int)
	scores["Alice"] = 95
	scores["Bob"] = 87
	fmt.Printf("make 创建: %v\n", scores)

	// 字面量创建
	config := map[string]string{
		"host": "localhost",
		"port": "8080",
		"env":  "dev",
	}
	fmt.Printf("literal 创建: %v\n", config)

	// nil map（不能写入）
	var nilMap map[string]int
	fmt.Printf("nil map: %v (isNil=%t)\n", nilMap, nilMap == nil)
	// nilMap["key"] = 1 // panic: assignment to entry in nil map
}

// demoMapCRUD 演示 map 的增删改查
func demoMapCRUD() {
	fmt.Println("\n=== Map CRUD ===")

	m := make(map[string]int)

	// Create / Update
	m["apple"] = 5
	m["banana"] = 3
	m["orange"] = 8
	fmt.Printf("添加后: %v\n", m)

	// Update
	m["apple"] = 10
	fmt.Printf("修改 apple=10: %v\n", m)

	// Read
	v := m["apple"]
	fmt.Printf("读取 apple: %d\n", v)

	// Delete
	delete(m, "banana")
	fmt.Printf("删除 banana 后: %v\n", m)

	// Comma ok idiom（检查 key 是否存在）
	value, ok := m["banana"]
	if !ok {
		fmt.Println("banana 不存在")
	} else {
		fmt.Printf("banana = %d\n", value)
	}

	value2, ok2 := m["apple"]
	fmt.Printf("apple: value=%d, exists=%t\n", value2, ok2)
}

// ------------------------------------------------------------
// 2. Map 迭代
// ------------------------------------------------------------

// demoMapIteration 演示 map 迭代（顺序不保证）
func demoMapIteration() {
	fmt.Println("\n=== Map 迭代 ===")

	m := map[string]int{
		"a": 1, "b": 2, "c": 3, "d": 4, "e": 5,
	}

	fmt.Println("迭代顺序不固定（多次运行可能不同）:")
	for k, v := range m {
		fmt.Printf("  %s -> %d\n", k, v)
	}

	// 只遍历 key
	fmt.Print("只遍历 key: ")
	for k := range m {
		fmt.Printf("%s ", k)
	}
	fmt.Println()

	// 按 key 排序遍历（保证输出顺序）
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Print("排序后遍历: ")
	for _, k := range keys {
		fmt.Printf("%s:%d ", k, m[k])
	}
	fmt.Println()
}

// ------------------------------------------------------------
// 3. Map 作为 Set
// ------------------------------------------------------------

// demoMapAsSet 演示使用 map 实现 set
func demoMapAsSet() {
	fmt.Println("\n=== Map 作为 Set ===")

	// 用 map[string]struct{} 实现 set（struct{} 不占空间）
	set := make(map[string]struct{})
	set["apple"] = struct{}{}
	set["banana"] = struct{}{}
	set["orange"] = struct{}{}

	// 添加重复元素不会报错
	set["apple"] = struct{}{}

	// 检查存在性
	_, ok := set["grape"]
	fmt.Printf("'grape' in set: %t\n", ok)
	_, ok2 := set["apple"]
	fmt.Printf("'apple' in set: %t\n", ok2)

	// 遍历 set
	fmt.Print("set 元素: ")
	for k := range set {
		fmt.Printf("%s ", k)
	}
	fmt.Println()

	// 删除
	delete(set, "banana")
	fmt.Printf("删除 banana 后: %v\n", set)
}

// ------------------------------------------------------------
// 4. Map 的 Struct Key
// ------------------------------------------------------------

type Point struct {
	X, Y int
}

// demoStructKey 演示使用结构体作为 map 的键
func demoStructKey() {
	fmt.Println("\n=== Struct 作为 Map Key ===")

	distances := map[Point]float64{
		{X: 0, Y: 0}: 0.0,
		{X: 3, Y: 4}: 5.0,
		{X: 1, Y: 1}: 1.414,
	}

	p := Point{X: 3, Y: 4}
	fmt.Printf("Point%v 距离: %.1f\n", p, distances[p])

	// 只有所有字段都可比较的结构体才能作为 key
	// 包含 slice/map 的结构体不能作为 key
	distances[Point{X: 5, Y: 12}] = 13.0
	fmt.Printf("distances 总数: %d\n", len(distances))
}

// ------------------------------------------------------------
// 5. 实战：词频统计
// ------------------------------------------------------------

// demoWordFreq 演示词频统计器
func demoWordFreq() {
	fmt.Println("\n=== 实战：词频统计 ===")

	text := "The quick brown fox jumps over the lazy dog. The dog barks, and the fox runs away."
	text = strings.ToLower(text)

	// 简单分词：按非字母字符分割
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	})

	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++
	}

	// 按频率排序输出
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range freq {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	fmt.Println("词频统计（按频率降序）:")
	for _, pair := range sorted {
		fmt.Printf("  %-10s %d\n", pair.Key+":", pair.Value)
	}
	fmt.Printf("总单词数: %d, 唯一单词数: %d\n", len(words), len(freq))
}

// ------------------------------------------------------------
// main
// ------------------------------------------------------------

func main() {
	fmt.Println("========================================")
	fmt.Println("  02-datastructures / map")
	fmt.Println("========================================")

	demoMapCreate()
	demoMapCRUD()
	demoMapIteration()
	demoMapAsSet()
	demoStructKey()
	demoWordFreq()

	fmt.Println("\n=== 完成 ===")
}