// trace 执行轨迹可视化示例
// 运行方式:
//
//	go run main.go           # 生成 trace.out
//	go tool trace trace.out  # 在浏览器中打开可视化界面
//
// 浏览器中可以看到:
//   - Schedule: goroutine 调度时间线
//   - User-defined tasks & regions: 自定义任务区间
//   - Network blocking profile: 网络阻塞
//   - Synchronization blocking: 同步阻塞
//   - Syscall blocking: 系统调用阻塞
//   - GC: 垃圾回收活动
//   - Goroutine analysis: 各 goroutine 统计
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

func main() {
	// 创建 trace 输出文件
	f, err := os.Create("trace.out")
	if err != nil {
		panic(fmt.Sprintf("创建 trace 文件失败: %v", err))
	}
	defer f.Close()

	// 启动 trace
	if err := trace.Start(f); err != nil {
		panic(fmt.Sprintf("启动 trace 失败: %v", err))
	}
	defer trace.Stop()

	ctx := context.Background()

	fmt.Println("trace 记录中, 请等待几秒...")
	fmt.Println("运行完成后: go tool trace trace.out")

	// 创建停止信号通道
	done := make(chan struct{})

	// ============================================================
	// 1. goroutine 创建与销毁
	// ============================================================
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		id := i
		go func() {
			defer wg.Done()
			// 用 trace region 标记 goroutine 活动区间
			region := trace.StartRegion(ctx, fmt.Sprintf("worker-%d", id))
			defer region.End()

			// 模拟计算工作
			_ = busyWork(100 + id*50)

			// 模拟 IO 等待（通过 sleep 模拟）
			time.Sleep(time.Duration(10+id*5) * time.Millisecond)
		}()
	}

	// ============================================================
	// 2. channel 操作（无缓冲通信 ping-pong）
	// ============================================================
	wg.Add(1)
	go func() {
		defer wg.Done()
		region := trace.StartRegion(ctx, "channel-ping-pong")
		defer region.End()

		ping := make(chan int)
		pong := make(chan int)

		// ping 发送者
		go func() {
			for i := 0; i < 20; i++ {
				ping <- i
				<-pong
			}
		}()

		// pong 接收者/发送者
		for i := 0; i < 20; i++ {
			val := <-ping
			_ = val
			pong <- val
		}
	}()

	// ============================================================
	// 3. GC 压力：频繁分配大量对象，观察 GC 活动
	// ============================================================
	wg.Add(1)
	go func() {
		defer wg.Done()
		region := trace.StartRegion(ctx, "gc-pressure")
		defer region.End()

		for i := 0; i < 50; i++ {
			_ = allocateHeap(5000) // 每次分配 5000 个对象
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// ============================================================
	// 4. 扇出模式：一个 producer 多个 consumer
	// ============================================================
	wg.Add(1)
	go func() {
		defer wg.Done()
		region := trace.StartRegion(ctx, "fan-out")
		defer region.End()

		jobs := make(chan int, 100)
		results := make(chan int, 100)

		// 3 个 worker 消费者
		for i := 0; i < 3; i++ {
			go func(id int) {
				for job := range jobs {
					// 模拟处理
					result := job * job
					results <- result
					_ = id
				}
			}(i)
		}

		// producer 发送 30 个任务
		for i := 0; i < 30; i++ {
			jobs <- i
		}
		close(jobs)

		// 收集结果
		for i := 0; i < 30; i++ {
			<-results
		}
	}()

	// ============================================================
	// 5. 用户等待，同时让 trace 积累足够数据
	// ============================================================
	go func() {
		// 先让前面的 goroutine 跑一会儿
		wg.Wait()
		// 额外再来一轮 GC 活动
		for i := 0; i < 20; i++ {
			_ = allocateHeap(2000)
			time.Sleep(1 * time.Millisecond)
		}
		close(done)
	}()

	<-done
	fmt.Println("\ntrace 完成！运行: go tool trace trace.out")
}

// busyWork 执行一些 CPU 计算
func busyWork(n int) int {
	result := 0
	for i := 0; i < n; i++ {
		result += i * i
		// 加入随机性避免编译器优化
		result ^= rand.Intn(100)
	}
	return result
}

// allocateHeap 在堆上分配大量对象，触发 GC
func allocateHeap(count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		// fmt.Sprintf 会导致逃逸到堆
		result[i] = fmt.Sprintf("data-%d-%d", i, rand.Intn(1000))
	}
	return result
}