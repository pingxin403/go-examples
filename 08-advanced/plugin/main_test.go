package main

import (
	"testing"
)

// ============ PluginManager 测试 ============

// TestNewPluginManager 测试插件管理器的创建
func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager()
	if pm == nil {
		t.Fatal("NewPluginManager() 返回 nil")
	}
	if pm.plugins == nil {
		t.Fatal("插件 map 未初始化")
	}
}

// TestPluginManager_RegisterAndGet 测试插件的注册和获取
func TestPluginManager_RegisterAndGet(t *testing.T) {
	pm := NewPluginManager()
	p := &UppercasePlugin{prefix: "[TEST] "}
	err := pm.Register(p)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	got, ok := pm.Get("uppercase")
	if !ok {
		t.Fatal("获取已注册插件返回 false")
	}
	if got != p {
		t.Error("获取到的插件与注册的不一致")
	}
}

// TestPluginManager_RegisterDuplicate 测试注册同名插件
func TestPluginManager_RegisterDuplicate(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&UppercasePlugin{prefix: "[A] "})
	err := pm.Register(&UppercasePlugin{prefix: "[B] "})
	if err == nil {
		t.Fatal("重复注册应返回错误")
	}
}

// TestPluginManager_GetNonExistent 测试获取不存在的插件
func TestPluginManager_GetNonExistent(t *testing.T) {
	pm := NewPluginManager()
	_, ok := pm.Get("nonexistent")
	if ok {
		t.Fatal("获取不存在的插件应返回 false")
	}
}

// TestPluginManager_ListPlugins 测试列出所有插件
func TestPluginManager_ListPlugins(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&UppercasePlugin{})
	pm.Register(&ReversePlugin{})
	pm.Register(&WordCountPlugin{})

	names := pm.ListPlugins()
	if len(names) != 3 {
		t.Fatalf("插件数量应为 3，实际为 %d", len(names))
	}

	// 验证所有插件名称都在列表中
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["uppercase"] || !nameSet["reverse"] || !nameSet["wordcount"] {
		t.Errorf("插件列表不完整: %v", names)
	}
}

// TestPluginManager_ListPlugins_Empty 测试空管理器的插件列表
func TestPluginManager_ListPlugins_Empty(t *testing.T) {
	pm := NewPluginManager()
	names := pm.ListPlugins()
	if names == nil {
		t.Fatal("空管理器的 ListPlugins 不应返回 nil")
	}
	if len(names) != 0 {
		t.Errorf("空管理器应返回空列表，实际为 %v", names)
	}
}

// TestPluginManager_ExecuteAll 测试顺序执行所有插件
func TestPluginManager_ExecuteAll(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&UppercasePlugin{prefix: ""})
	pm.Register(&ReversePlugin{})
	pm.Register(&WordCountPlugin{})

	results := pm.ExecuteAll("hello world")

	if len(results) != 3 {
		t.Fatalf("结果数应为 3，实际为 %d", len(results))
	}

	// 验证 uppercase 结果
	upper, ok := results["uppercase"]
	if !ok {
		t.Fatal("缺少 uppercase 结果")
	}
	if upper != "HELLO WORLD" {
		t.Errorf("uppercase 结果应为 'HELLO WORLD'，实际为 %v", upper)
	}

	// 验证 reverse 结果
	rev, ok := results["reverse"]
	if !ok {
		t.Fatal("缺少 reverse 结果")
	}
	if rev != "dlrow olleh" {
		t.Errorf("reverse 结果应为 'dlrow olleh'，实际为 %v", rev)
	}

	// 验证 wordcount 结果
	wc, ok := results["wordcount"]
	if !ok {
		t.Fatal("缺少 wordcount 结果")
	}
	wcMap, ok := wc.(map[string]int)
	if !ok {
		t.Fatalf("wordcount 结果类型应为 map[string]int，实际为 %T", wc)
	}
	if wcMap["hello"] != 1 || wcMap["world"] != 1 {
		t.Errorf("wordcount 结果不正确: %v", wcMap)
	}
}

// ============ UppercasePlugin 测试 ============

// TestUppercasePlugin_Name 测试插件名称
func TestUppercasePlugin_Name(t *testing.T) {
	p := &UppercasePlugin{}
	if p.Name() != "uppercase" {
		t.Errorf("名称应为 'uppercase'，实际为 %q", p.Name())
	}
}

// TestUppercasePlugin_Execute 测试大写转换
func TestUppercasePlugin_Execute(t *testing.T) {
	p := &UppercasePlugin{}
	result, err := p.Execute("hello")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("结果应为 'HELLO'，实际为 %v", result)
	}
}

// TestUppercasePlugin_ExecuteWithPrefix 测试带前缀的大写转换
func TestUppercasePlugin_ExecuteWithPrefix(t *testing.T) {
	p := &UppercasePlugin{prefix: "[APP] "}
	result, err := p.Execute("test")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	if result != "[APP] TEST" {
		t.Errorf("结果应为 '[APP] TEST'，实际为 %v", result)
	}
}

// TestUppercasePlugin_Execute_NonStringInput 测试非字符串输入
func TestUppercasePlugin_Execute_NonStringInput(t *testing.T) {
	p := &UppercasePlugin{}
	_, err := p.Execute(123)
	if err == nil {
		t.Fatal("非字符串输入应返回错误")
	}
}

// TestUppercasePlugin_Init 测试初始化
func TestUppercasePlugin_Init(t *testing.T) {
	p := &UppercasePlugin{}
	err := p.Init(map[string]interface{}{"prefix": "[INIT] "})
	if err != nil {
		t.Fatalf("Init 返回错误: %v", err)
	}
	if p.prefix != "[INIT] " {
		t.Errorf("prefix 应为 '[INIT] '，实际为 %q", p.prefix)
	}
}

// TestUppercasePlugin_Close 测试关闭
func TestUppercasePlugin_Close(t *testing.T) {
	p := &UppercasePlugin{}
	err := p.Close()
	if err != nil {
		t.Fatalf("Close 返回错误: %v", err)
	}
}

// ============ ReversePlugin 测试 ============

// TestReversePlugin_Name 测试插件名称
func TestReversePlugin_Name(t *testing.T) {
	p := &ReversePlugin{}
	if p.Name() != "reverse" {
		t.Errorf("名称应为 'reverse'，实际为 %q", p.Name())
	}
}

// TestReversePlugin_Execute 测试字符串反转
func TestReversePlugin_Execute(t *testing.T) {
	p := &ReversePlugin{}
	result, err := p.Execute("hello")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	if result != "olleh" {
		t.Errorf("结果应为 'olleh'，实际为 %v", result)
	}
}

// TestReversePlugin_Execute_Empty 测试空字符串反转
func TestReversePlugin_Execute_Empty(t *testing.T) {
	p := &ReversePlugin{}
	result, err := p.Execute("")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	if result != "" {
		t.Errorf("空字符串反转应返回空，实际为 %v", result)
	}
}

// TestReversePlugin_Execute_Unicode 测试 Unicode 字符串反转
func TestReversePlugin_Execute_Unicode(t *testing.T) {
	p := &ReversePlugin{}
	result, err := p.Execute("世界")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	if result != "界世" {
		t.Errorf("结果应为 '界世'，实际为 %v", result)
	}
}

// TestReversePlugin_Execute_NonStringInput 测试非字符串输入
func TestReversePlugin_Execute_NonStringInput(t *testing.T) {
	p := &ReversePlugin{}
	_, err := p.Execute(3.14)
	if err == nil {
		t.Fatal("非字符串输入应返回错误")
	}
}

// TestReversePlugin_Execute_Palindrome 测试回文串反转
func TestReversePlugin_Execute_Palindrome(t *testing.T) {
	p := &ReversePlugin{}
	result, err := p.Execute("aba")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	if result != "aba" {
		t.Errorf("回文反转结果应为自身，实际为 %v", result)
	}
}

// TestReversePlugin_Init 测试初始化（空实现）
func TestReversePlugin_Init(t *testing.T) {
	p := &ReversePlugin{}
	err := p.Init(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Init 返回错误: %v", err)
	}
}

// TestReversePlugin_Close 测试关闭（空实现）
func TestReversePlugin_Close(t *testing.T) {
	p := &ReversePlugin{}
	err := p.Close()
	if err != nil {
		t.Fatalf("Close 返回错误: %v", err)
	}
}

// ============ WordCountPlugin 测试 ============

// TestWordCountPlugin_Name 测试插件名称
func TestWordCountPlugin_Name(t *testing.T) {
	p := &WordCountPlugin{}
	if p.Name() != "wordcount" {
		t.Errorf("名称应为 'wordcount'，实际为 %q", p.Name())
	}
}

// TestWordCountPlugin_Execute 测试词频统计
func TestWordCountPlugin_Execute(t *testing.T) {
	p := &WordCountPlugin{}
	result, err := p.Execute("hello world hello")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	counts, ok := result.(map[string]int)
	if !ok {
		t.Fatalf("结果类型应为 map[string]int，实际为 %T", result)
	}
	if counts["hello"] != 2 {
		t.Errorf("'hello' 计数应为 2，实际为 %d", counts["hello"])
	}
	if counts["world"] != 1 {
		t.Errorf("'world' 计数应为 1，实际为 %d", counts["world"])
	}
}

// TestWordCountPlugin_Execute_CaseSensitive 测试大小写敏感模式
func TestWordCountPlugin_Execute_CaseSensitive(t *testing.T) {
	p := &WordCountPlugin{caseSensitive: true}
	result, err := p.Execute("Hello hello")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	counts, ok := result.(map[string]int)
	if !ok {
		t.Fatalf("结果类型应为 map[string]int，实际为 %T", result)
	}
	if counts["Hello"] != 1 || counts["hello"] != 1 {
		t.Errorf("大小写应区分: Hello=%d, hello=%d", counts["Hello"], counts["hello"])
	}
}

// TestWordCountPlugin_Execute_CaseInsensitive 测试大小写不敏感模式
func TestWordCountPlugin_Execute_CaseInsensitive(t *testing.T) {
	p := &WordCountPlugin{caseSensitive: false}
	result, err := p.Execute("Hello hello")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	counts, ok := result.(map[string]int)
	if !ok {
		t.Fatalf("结果类型应为 map[string]int，实际为 %T", result)
	}
	if counts["hello"] != 2 {
		t.Errorf("忽略大小写时 'hello' 计数应为 2，实际为 %d", counts["hello"])
	}
}

// TestWordCountPlugin_Execute_EmptyInput 测试空字符串
func TestWordCountPlugin_Execute_EmptyInput(t *testing.T) {
	p := &WordCountPlugin{}
	result, err := p.Execute("")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	counts, ok := result.(map[string]int)
	if !ok {
		t.Fatalf("结果类型应为 map[string]int，实际为 %T", result)
	}
	if len(counts) != 0 {
		t.Errorf("空字符串应返回空 map，实际为 %v", counts)
	}
}

// TestWordCountPlugin_Execute_WithNumbers 测试含数字的单词
func TestWordCountPlugin_Execute_WithNumbers(t *testing.T) {
	p := &WordCountPlugin{}
	result, err := p.Execute("test123 test456")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	counts, ok := result.(map[string]int)
	if !ok {
		t.Fatalf("结果类型应为 map[string]int，实际为 %T", result)
	}
	if counts["test123"] != 1 || counts["test456"] != 1 {
		t.Errorf("含数字的单词计数不正确: %v", counts)
	}
}

// TestWordCountPlugin_Execute_WithPunctuation 测试标点符号分隔
func TestWordCountPlugin_Execute_WithPunctuation(t *testing.T) {
	p := &WordCountPlugin{}
	result, err := p.Execute("hello,world!hello")
	if err != nil {
		t.Fatalf("Execute 返回错误: %v", err)
	}
	counts, ok := result.(map[string]int)
	if !ok {
		t.Fatalf("结果类型应为 map[string]int，实际为 %T", result)
	}
	if counts["hello"] != 2 || counts["world"] != 1 {
		t.Errorf("标点分隔的单词计数不正确: %v", counts)
	}
}

// TestWordCountPlugin_Execute_NonStringInput 测试非字符串输入
func TestWordCountPlugin_Execute_NonStringInput(t *testing.T) {
	p := &WordCountPlugin{}
	_, err := p.Execute(true)
	if err == nil {
		t.Fatal("非字符串输入应返回错误")
	}
}

// TestWordCountPlugin_Init 测试初始化
func TestWordCountPlugin_Init(t *testing.T) {
	p := &WordCountPlugin{}
	err := p.Init(map[string]interface{}{"case_sensitive": true})
	if err != nil {
		t.Fatalf("Init 返回错误: %v", err)
	}
	if !p.caseSensitive {
		t.Error("caseSensitive 应被设为 true")
	}
}

// TestWordCountPlugin_Init_Default 测试默认初始化
func TestWordCountPlugin_Init_Default(t *testing.T) {
	p := &WordCountPlugin{}
	err := p.Init(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Init 返回错误: %v", err)
	}
	if p.caseSensitive {
		// caseSensitive 默认为 false
		t.Error("默认 caseSensitive 应为 false")
	}
}

// TestWordCountPlugin_Close 测试关闭
func TestWordCountPlugin_Close(t *testing.T) {
	p := &WordCountPlugin{}
	err := p.Close()
	if err != nil {
		t.Fatalf("Close 返回错误: %v", err)
	}
}

// ============ 组合 Pipeline 测试 ============

// TestPluginPipeline 测试多个插件组合使用
func TestPluginPipeline(t *testing.T) {
	reverse := &ReversePlugin{}
	uppercase := &UppercasePlugin{prefix: ""}

	step1, err := reverse.Execute("Go is Awesome")
	if err != nil {
		t.Fatalf("reverse.Execute 错误: %v", err)
	}
	step2, err := uppercase.Execute(step1)
	if err != nil {
		t.Fatalf("uppercase.Execute 错误: %v", err)
	}
	// "Go is Awesome" 反转 = "emosewA si oG", 大写 = "EMOSEWA SI OG"
	if step2 != "EMOSEWA SI OG" {
		t.Errorf("流水线结果应为 'EMOSEWA SI OG'，实际为 %v", step2)
	}
}

// ============ loadPluginFromFile 测试 ============

// TestLoadPluginFromFile_NonExistent 测试加载不存在的 .so 文件
func TestLoadPluginFromFile_NonExistent(t *testing.T) {
	_, err := loadPluginFromFile("/nonexistent/plugin.so")
	if err == nil {
		t.Fatal("加载不存在的文件应返回错误")
	}
}

// ============ 接口兼容性测试 ============

// TestPluginInterface_CompileCheck 测试插件实现是否正确（编译期验证）
func TestPluginInterface_CompileCheck(t *testing.T) {
	// 编译期验证: UppercasePlugin 实现了 Plugin 接口
	var _ Plugin = (*UppercasePlugin)(nil)
	// 编译期验证: ReversePlugin 实现了 Plugin 接口
	var _ Plugin = (*ReversePlugin)(nil)
	// 编译期验证: WordCountPlugin 实现了 Plugin 接口
	var _ Plugin = (*WordCountPlugin)(nil)
}