package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

/*
 * ===========================================
 *  PGO（Profile Guided Optimization）示例
 * ===========================================
 *
 * PGO 是 Go 1.20 引入的配置文件引导优化，Go 1.21 默认开启。
 * 它允许编译器根据运行时 profiling 数据，优化 hot path：
 *   - 内联优化（inline）
 *   - 代码布局优化
 *   - 分支预测优化
 *   - 虚函数去虚拟化（devirtualization）
 *
 * PGO 工作流：
 *   1. 构建不带 PGO 的可执行文件（作为 baseline）
 *      $ go build -o without-pgo .
 *
 *   2. 运行程序生成 CPU profile
 *      $ go test -bench=. -cpuprofile default.pgo
 *      或直接运行程序：
 *      $ go run .
 *      然后把生成的 default.pgo 复制到项目根目录
 *
 *   3. 构建带 PGO 的可执行文件
 *      $ go build -pgo=auto -o with-pgo .
 *      # Go 1.21+ 会自动查找根目录的 default.pgo
 *
 *   4. 比较两者的性能差异
 *      $ time ./without-pgo
 *      $ time ./with-pgo
 *
 * 注意：本示例主要是教学演示，实际的 PGO 收益需要在
 * 真实工作负载上运行 profiling 才能体现。
 */

// ============ Hot Path：数据集排序与搜索 ============

// generateDataset 生成随机整数数据集（模拟真实工作负载）
func generateDataset(size int) []int {
	data := make([]int, size)
	for i := range data {
		data[i] = rand.Intn(1000000)
	}
	return data
}

// processDatasetHotPath 模拟排序、搜索、聚合的密集计算路径
// 这将被 PGO 识别为 hot path，进行重点优化
func processDatasetHotPath(data []int) (sorted []int, sum int64, avg float64, median int) {
	// 复制一份避免修改原始数据
	sorted = make([]int, len(data))
	copy(sorted, data)

	// 排序（Go 1.22 使用 pdqsort，但大数组排序仍然是 hot path）
	sort.Ints(sorted)

	// 求和
	for _, v := range sorted {
		sum += int64(v)
	}

	// 平均值
	avg = float64(sum) / float64(len(sorted))

	// 中位数
	n := len(sorted)
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	} else {
		median = sorted[n/2]
	}

	return
}

// ============ Hot Path：字符串处理 ============

// textProcessor 模拟文本处理 pipeline
type textProcessor struct {
	buffer []string
}

// newTextProcessor 创建新的文本处理器
func newTextProcessor() *textProcessor {
	return &textProcessor{
		buffer: make([]string, 0, 1000),
	}
}

// ingest 接收文本行
func (tp *textProcessor) ingest(lines []string) {
	tp.buffer = append(tp.buffer, lines...)
}

// analyze 分析文本数据（hot path）
func (tp *textProcessor) analyze() map[string]int {
	result := make(map[string]int)
	for _, line := range tp.buffer {
		// 模拟一些处理
		for i := 0; i < len(line); i++ {
			if line[i] >= 'a' && line[i] <= 'z' {
				// 统计首字母分布（模拟真实处理逻辑）
				key := string(line[i])
				result[key]++
				break
			}
		}
	}
	return result
}

// ============ Benchmark-worthy 工作负载 ============

// DataWorkload 可被 go test -bench 采集的性能负载
type DataWorkload struct {
	datasets [][]int
}

// NewDataWorkload 创建一个测试负载
func NewDataWorkload(count, size int) *DataWorkload {
	dw := &DataWorkload{
		datasets: make([][]int, count),
	}
	for i := 0; i < count; i++ {
		dw.datasets[i] = generateDataset(size)
	}
	return dw
}

// RunWorkload 运行主要工作负载
func (dw *DataWorkload) RunWorkload() (totalSum int64) {
	for _, data := range dw.datasets {
		_, sum, _, _ := processDatasetHotPath(data)
		totalSum += sum
	}
	return
}

func main() {
	fmt.Println("================================================")
	fmt.Println("  PGO（Profile Guided Optimization）示例")
	fmt.Println("================================================")

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	fmt.Println("\n--- 生成测试数据 ---")
	data := generateDataset(500000)
	fmt.Printf("生成 %d 个随机整数\n", len(data))

	// ====================================================
	// Hot Path 1：排序和统计分析
	// ====================================================
	fmt.Println("\n--- Hot Path 1: 数据集排序与统计 ---")

	start := time.Now()
	_, sum, avg, median := processDatasetHotPath(data)
	elapsed := time.Since(start)

	fmt.Printf("排序和统计耗时: %v\n", elapsed)
	fmt.Printf("  总和: %d\n", sum)
	fmt.Printf("  平均: %.2f\n", avg)
	fmt.Printf("  中位数: %d\n", median)

	// ====================================================
	// Hot Path 2：文本处理
	// ====================================================
	fmt.Println("\n--- Hot Path 2: 文本处理 ---")

	tp := newTextProcessor()
	tp.ingest([]string{
		"apple banana cherry",
		"date elderberry fig",
		"grape honeydew iilama",
		"jackfruit kiwi lemon",
		"mango nectarine orange",
		"papaya quince raspberry",
		"strawberry tangerine ugli",
		"vanilla watermelon xigua",
		"yam zucchini",
	})

	result := tp.analyze()
	fmt.Println("文本分析结果（首字母分布）:")
	for k, v := range result {
		fmt.Printf("  '%s': %d\n", k, v)
	}

	// ====================================================
	// Hot Path 3：批量负载（Benchmark 模拟）
	// ====================================================
	fmt.Println("\n--- Hot Path 3: 批量工作负载 ---")

	workload := NewDataWorkload(10, 100000)
	start = time.Now()
	totalSum := workload.RunWorkload()
	elapsed = time.Since(start)

	fmt.Printf("批量处理 10 组数据，每组 100000 个元素\n")
	fmt.Printf("总耗时: %v\n", elapsed)
	fmt.Printf("总和: %d\n", totalSum)

	// ====================================================
	// PGO 工作流说明
	// ====================================================
	fmt.Println("\n================================================")
	fmt.Println("  PGO 使用工作流")
	fmt.Println("================================================")
	fmt.Println(`
PGO 工作流（Profile Guided Optimization）：

1️⃣  基准构建（不带 PGO）：
    $ go build -o without-pgo .

2️⃣  收集 Profile：
    # 方法 A：使用 test bench（推荐）
    $ go test -bench=. -cpuprofile default.pgo
    # 注意：需要在 *_test.go 中编写与 main 相同逻辑的 Benchmark 函数

    # 方法 B：直接运行程序（Go 1.21+）
    $ go run .  # 运行后会在当前目录生成 default.pgo

3️⃣  PGO 构建：
    $ go build -pgo=auto -o with-pgo .
    # Go 1.21+ 自动查找 default.pgo，无需手动指定

    # 或者显式指定 profile 文件：
    $ go build -pgo=default.pgo -o with-pgo .

4️⃣  性能对比：
    $ hyperfine ./without-pgo ./with-pgo
    # 或
    $ time ./without-pgo
    $ time ./with-pgo

5️⃣  注意事项：
    • Profile 应反映真实生产负载（而非合成测试数据）
    • 应用代码大幅更新后应重新采集 profile
    • 不同部署环境可能需要不同的 profile
    • PGO 优化效果通常在 2-15% 之间（取决于应用特性）
    • 虚函数调用多的代码（大量 interface 方法）收益最大`)
}