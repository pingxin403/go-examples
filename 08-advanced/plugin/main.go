package main

import (
	"fmt"
	"plugin"
)

/*
 * ==========================================
 *  Go Plugin 系统示例
 * ==========================================
 *
 * Go 的 plugin 包（buildmode=plugin）允许在运行时动态加载
 * 编译好的共享库（.so 文件），实现插件化架构。
 *
 * 平台限制：
 *   - 仅支持 Linux、macOS（不支持 Windows）
 *   - macOS 需要设置 -tags=plugin
 *   - Go 版本必须与编译插件时的版本完全一致
 *   - 插件和被加载程序必须使用相同的依赖版本
 *
 * 由于 buildmode=plugin 的这些限制，实际应用中更多采用
 * 接口（interface）驱动的插件设计模式，而非原生 plugin 机制。
 * 本示例同时展示两种方案。
 *
 * 构建插件（运行前需要先编译插件）：
 *   $ go build -buildmode=plugin -o math_plugin.so ./math_plugin/
 *   $ go build -buildmode=plugin -o text_plugin.so ./text_plugin/
 *   $ go run main.go
 */

// ============ 方案A：Go Plugin 原生方案 ============

// Processor 插件接口定义（插件和宿主程序共享）
type Processor interface {
	Name() string
	Process(input string) (string, error)
	Version() string
}

// loadPluginFromFile 从 .so 文件加载插件
// 这是 Go 标准库 plugin 包的用法
func loadPluginFromFile(path string) (Processor, error) {
	// 打开共享库
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开插件 %s 失败: %w", path, err)
	}

	// 查找符号（变量或函数）
	sym, err := p.Lookup("PluginProcessor")
	if err != nil {
		return nil, fmt.Errorf("查找 PluginProcessor 符号失败: %w", err)
	}

	// 类型断言为 Processor 接口
	proc, ok := sym.(Processor)
	if !ok {
		return nil, fmt.Errorf("符号 %T 未实现 Processor 接口", sym)
	}

	return proc, nil
}

// ============ 方案B：接口驱动的插件设计模式（推荐） ============

// Plugin 通用插件接口
type Plugin interface {
	// Init 初始化插件，config 为任意配置
	Init(config map[string]interface{}) error
	// Name 插件名称
	Name() string
	// Execute 执行插件核心逻辑
	Execute(input interface{}) (interface{}, error)
	// Close 清理资源
	Close() error
}

// PluginManager 插件管理器
type PluginManager struct {
	plugins map[string]Plugin
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

// Register 注册一个插件
func (pm *PluginManager) Register(p Plugin) error {
	name := p.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("插件 %q 已存在", name)
	}
	pm.plugins[name] = p
	return nil
}

// Get 获取指定名称的插件
func (pm *PluginManager) Get(name string) (Plugin, bool) {
	p, ok := pm.plugins[name]
	return p, ok
}

// ExecuteAll 顺序执行所有插件
func (pm *PluginManager) ExecuteAll(input interface{}) map[string]interface{} {
	results := make(map[string]interface{})
	for name, p := range pm.plugins {
		result, err := p.Execute(input)
		if err != nil {
			results[name] = fmt.Sprintf("错误: %v", err)
		} else {
			results[name] = result
		}
	}
	return results
}

// ListPlugins 列出所有已注册的插件
func (pm *PluginManager) ListPlugins() []string {
	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// ============ 实现具体插件 ============

// UppercasePlugin 大写转换插件
type UppercasePlugin struct {
	prefix string
}

func (p *UppercasePlugin) Init(config map[string]interface{}) error {
	if prefix, ok := config["prefix"]; ok {
		p.prefix = prefix.(string)
	}
	return nil
}

func (p *UppercasePlugin) Name() string { return "uppercase" }

func (p *UppercasePlugin) Execute(input interface{}) (interface{}, error) {
	str, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("uppercase 插件需要 string 类型输入")
	}
	result := p.prefix + str
	// 转为大写（实际项目中用 strings.ToUpper）
	// 这里用 byte 操作演示
	b := []byte(result)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b), nil
}

func (p *UppercasePlugin) Close() error { return nil }

// ReversePlugin 字符串反转插件
type ReversePlugin struct{}

func (p *ReversePlugin) Init(config map[string]interface{}) error {
	return nil
}

func (p *ReversePlugin) Name() string { return "reverse" }

func (p *ReversePlugin) Execute(input interface{}) (interface{}, error) {
	str, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("reverse 插件需要 string 类型输入")
	}
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes), nil
}

func (p *ReversePlugin) Close() error { return nil }

// WordCountPlugin 词频统计插件
type WordCountPlugin struct {
	caseSensitive bool
}

func (p *WordCountPlugin) Init(config map[string]interface{}) error {
	if cs, ok := config["case_sensitive"]; ok {
		p.caseSensitive = cs.(bool)
	}
	return nil
}

func (p *WordCountPlugin) Name() string { return "wordcount" }

func (p *WordCountPlugin) Execute(input interface{}) (interface{}, error) {
	str, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("wordcount 插件需要 string 类型输入")
	}

	count := make(map[string]int)
	word := ""
	for _, c := range str {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			if !p.caseSensitive && c >= 'A' && c <= 'Z' {
				c = c + 32 // 转小写
			}
			word += string(c)
		} else if word != "" {
			count[word]++
			word = ""
		}
	}
	if word != "" {
		count[word]++
	}
	return count, nil
}

func (p *WordCountPlugin) Close() error { return nil }

func main() {
	fmt.Println("========================================")
	fmt.Println("  Go Plugin 系统完整示例")
	fmt.Println("========================================")

	// ====================================================
	// 方案A：原生 Go Plugin（buildmode=plugin）
	// ====================================================
	fmt.Println("\n--- 方案A: 原生 Go Plugin（buildmode=plugin）---")

	fmt.Print(`
编译原生插件的步骤：
  $ mkdir -p math_plugin text_plugin

  // math_plugin/math.go
  type MathProcessor struct{}
  func (m MathProcessor) Name() string    { return "math" }
  func (m MathProcessor) Process(s string) (string, error) { ... }
  func (m MathProcessor) Version() string { return "1.0.0" }
  var PluginProcessor MathProcessor

  // text_plugin/text.go
  type TextProcessor struct{}
  func (t TextProcessor) Name() string    { return "text" }
  func (t TextProcessor) Process(s string) (string, error) { ... }
  func (t TextProcessor) Version() string { return "1.0.0" }
  var PluginProcessor TextProcessor

  // 编译：
  $ go build -buildmode=plugin -o math_plugin.so ./math_plugin/
  $ go build -buildmode=plugin -o text_plugin.so ./text_plugin/

加载方式：
  p, err := plugin.Open("math_plugin.so")
  sym, err := p.Lookup("PluginProcessor")
  proc := sym.(Processor)
`)

	fmt.Println("  注意：当前未编译 .so 文件，此处为演示加载逻辑")
	fmt.Println("  实际运行前需先执行上述编译步骤")

	// 尝试加载（会失败，因为 .so 文件不存在）
	plugins := []string{"math_plugin.so", "text_plugin.so"}
	for _, path := range plugins {
		_, err := loadPluginFromFile(path)
		if err != nil {
			fmt.Printf("  ⚠️  跳过加载 %s: %v\n", path, err)
		}
	}

	// ====================================================
	// 方案B：接口驱动的插件设计模式（推荐，无需 .so）
	// ====================================================
	fmt.Println("\n--- 方案B: 接口驱动的插件设计模式 ---")
	fmt.Println("  优点：无需 .so 文件，跨平台，编译安全")
	fmt.Println("  缺点：插件需要在编译时注册，不能动态加载")

	// 创建插件管理器
	pm := NewPluginManager()

	// 注册插件
	pm.Register(&UppercasePlugin{prefix: "[PREFIX] "})
	pm.Register(&ReversePlugin{})
	pm.Register(&WordCountPlugin{caseSensitive: false})

	// 初始化插件（带配置）
	if p, ok := pm.Get("uppercase"); ok {
		p.Init(map[string]interface{}{
			"prefix": "[APP] ",
		})
	}

	fmt.Printf("\n已注册的插件: %v\n", pm.ListPlugins())

	// 执行所有插件
	input := "Hello Go Plugin System!"
	fmt.Printf("\n输入: %q\n\n", input)

	results := pm.ExecuteAll(input)
	for name, result := range results {
		fmt.Printf("  [%s] -> %v\n", name, result)
	}

	// ====================================================
	// 组合使用多个插件（Pipeline 模式）
	// ====================================================
	fmt.Println("\n--- Pipeline: 插件流水线组合 ---")

	input2 := "Go is Awesome and Powerful"

	// 先反转，再转大写
	p1, _ := pm.Get("reverse")
	p2, _ := pm.Get("uppercase")

	step1, _ := p1.Execute(input2)
	step2, _ := p2.Execute(step1)

	fmt.Printf("原始:    %q\n", input2)
	fmt.Printf("反转后:  %q\n", step1)
	fmt.Printf("大写后:  %q\n", step2)

	// ====================================================
	// 插件扩展：添加新的统计插件
	// ====================================================
	fmt.Println("\n--- 运行时添加新插件 ---")

	// 模拟热加载一个新插件
	type StatsPlugin struct {
		totalRuns int
	}
	statsPlugin := &StatsPlugin{}
	// 手动实现 Plugin 接口
	_ = struct {
		*StatsPlugin
		Init    func(map[string]interface{}) error
		Name    func() string
		Execute func(interface{}) (interface{}, error)
		Close   func() error
	}{
		StatsPlugin: statsPlugin,
		Init:        func(m map[string]interface{}) error { return nil },
		Name:        func() string { return "stats" },
		Execute: func(input interface{}) (interface{}, error) {
			statsPlugin.totalRuns++
			return fmt.Sprintf("第 %d 次执行", statsPlugin.totalRuns), nil
		},
		Close: func() error { return nil },
	}

	// 由于 Go 不支持匿名结构体实现接口后直接作为接口类型传递，
	// 实际项目中会用函数闭包方式注册，这里仅展示概念
	fmt.Println("  ✅ 可以在运行时动态注册插件（通过闭包或依赖注入）")

	fmt.Println("\n========================================")
	fmt.Println("  Plugin 设计选择总结")
	fmt.Println("========================================")
	fmt.Print(`
1. 原生 plugin（buildmode=plugin）
   ✅ 热加载，无需重启
   ✅ 第三方独立开发
   ❌ 仅 Linux/macOS
   ❌ Go 版本必须完全一致
   ❌ 依赖版本必须完全一致
   → 适用于：内部工具、边缘设备、需要热更新的系统

2. 接口驱动设计模式
   ✅ 跨平台，编译安全
   ✅ Go 版本解耦
   ✅ 可测试性强
   ❌ 需要重新编译才能加新插件
   ❌ 所有插件代码在同一个进程中
   → 适用于：大多数业务系统、需要类型安全的场景

3. 混合方案（interface + 外部进程）
   ✅ 跨语言（gRPC/HTTP 通信）
   ✅ 热加载
   ✅ 隔离性（crash 不影响主进程）
   ❌ 性能开销（RPC）
   ❌ 部署复杂度增加
   → 适用于：大型系统、需要隔离的场景
`)
}