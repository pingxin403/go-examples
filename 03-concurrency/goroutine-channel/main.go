package main

import (
	"fmt"
	"time"
)

// ============================================================
// goroutine + channel 基础示例
// 涵盖：goroutine 启动、无缓冲/有缓冲 channel、close/range、
// 生产者-消费者、channel 方向、工作池
// ============================================================

func main() {
	// ============================================================
	// 1. 基本 goroutine — 用 go 关键字启动并发任务
	// ============================================================
	fmt.Println("=== 1. 基本 goroutine ===")

	// goroutine 是轻量级线程，由 Go 运行时调度
	go func() {
		fmt.Println("  子 goroutine: 你好！")
	}()

	// 主 goroutine 若不等待，子 goroutine 可能来不及执行
	// 这里简单 sleep 一下让子 goroutine 跑完
	time.Sleep(10 * time.Millisecond)
	fmt.Println("  主 goroutine: 结束")

	// ============================================================
	// 2. 无缓冲 channel — 同步通信（发送和接收必须同时就绪）
	// ============================================================
	fmt.Println("\n=== 2. 无缓冲 channel（同步通信）===")

	// 无缓冲 channel：发送方会阻塞直到接收方读取，反之亦然
	ch := make(chan string)

	go func() {
		// 模拟一些工作
		time.Sleep(50 * time.Millisecond)
		ch <- "来自子 goroutine 的消息" // 发送，阻塞直到 main 接收
		fmt.Println("  子 goroutine: 消息已发送")
	}()

	// main goroutine 接收（会阻塞直到数据到达）
	msg := <-ch
	fmt.Printf("  主 goroutine: 收到 -> %q\n", msg)

	// ============================================================
	// 3. 有缓冲 channel — 异步通信
	// ============================================================
	fmt.Println("\n=== 3. 有缓冲 channel（异步通信）===")

	// 有缓冲 channel：缓冲区未满时发送不阻塞，未空时接收不阻塞
	bufCh := make(chan int, 3)

	// 可以连续发送 3 个值而不阻塞（缓冲区大小=3）
	bufCh <- 10
	bufCh <- 20
	bufCh <- 30
	fmt.Println("  已发送 3 个值到有缓冲 channel")

	// 第 4 个发送会阻塞（缓冲区已满）
	go func() {
		time.Sleep(50 * time.Millisecond)
		bufCh <- 40
		fmt.Println("  第 4 个值已发送（阻塞后写入）")
	}()

	// 读取一个值，为第 4 个值腾出空间
	fmt.Printf("  读取: %d\n", <-bufCh)
	time.Sleep(100 * time.Millisecond)

	// 读取剩余值
	close(bufCh) // 先关闭，然后 range
	// 注意：上面已经 close 了，下面不能用 range 再读，所以手动读剩下的
	// 重新演示 — 用 range 遍历
	_ = bufCh // 忽略，下面重新演示 range

	// 重新演示 range over channel
	fmt.Println("\n  --- range over channel ---")
	// 新 channel 避免混乱
	rangeCh := make(chan int, 3)
	rangeCh <- 1
	rangeCh <- 2
	rangeCh <- 3
	close(rangeCh) // 必须关闭，否则 range 永远阻塞

	for val := range rangeCh {
		fmt.Printf("  range 读取: %d\n", val)
	}

	// ============================================================
	// 4. 关闭 channel 和 range 遍历
	// ============================================================
	fmt.Println("\n=== 4. 关闭 channel + range 遍历 ===")

	gen := func(count int) chan int {
		ch := make(chan int)
		go func() {
			for i := 1; i <= count; i++ {
				ch <- i * 10
			}
			close(ch) // 发送完成后关闭 channel
		}()
		return ch
	}

	for v := range gen(3) {
		fmt.Printf("  从生成器读取: %d\n", v)
	}

	// ============================================================
	// 5. 生产者-消费者模式
	// ============================================================
	fmt.Println("\n=== 5. 生产者-消费者模式 ===")

	jobs := make(chan int, 5)
	done := make(chan bool)

	// 消费者
	go func() {
		for job := range jobs {
			fmt.Printf("  消费者处理: job %d\n", job)
			time.Sleep(20 * time.Millisecond)
		}
		done <- true // 通知主 goroutine 消费完成
		fmt.Println("  消费者退出")
	}()

	// 生产者
	for i := 1; i <= 5; i++ {
		jobs <- i
		fmt.Printf("  生产者发布: job %d\n", i)
	}
	close(jobs) // 关闭通知消费者没有更多 job

	<-done // 等待消费者完成

	// ============================================================
	// 6. Channel 方向 — 函数参数指定 send-only / receive-only
	// ============================================================
	fmt.Println("\n=== 6. Channel 方向 ===")

	// sendOnly 参数：chan<- int 表示只能发送
	// recvOnly 参数：<-chan int 表示只能接收
	sendTask := func(out chan<- string, id int) {
		out <- fmt.Sprintf("任务 %d 完成", id)
		// out <- "ok" // 编译错误：只能发送
	}

	recvTask := func(in <-chan string) int {
		msg := <-in
		fmt.Printf("  收到: %s\n", msg)
		return len(msg)
		// in <- "x" // 编译错误：只能接收
	}

	taskCh := make(chan string, 3)
	for i := 1; i <= 3; i++ {
		sendTask(taskCh, i)
	}
	// 用方向明确的函数读取
	taskCh2 := taskCh // 复用
	for i := 1; i <= 3; i++ {
		recvTask(taskCh2)
	}
	_ = taskCh // 避免未使用警告

	// ============================================================
	// 7. Worker Pool 模式 — 固定数量 worker 处理任务
	// ============================================================
	fmt.Println("\n=== 7. Worker Pool 模式 ===")

	const numWorkers = 3
	const numJobs = 10

	workerJobs := make(chan int, numJobs)
	workerResults := make(chan int, numJobs)

	// 启动 worker
	for w := 1; w <= numWorkers; w++ {
		go worker(w, workerJobs, workerResults) // worker 定义为独立函数
	}

	// 发送任务
	for j := 1; j <= numJobs; j++ {
		workerJobs <- j
	}
	close(workerJobs)

	// 收集结果
	for r := 1; r <= numJobs; r++ {
		result := <-workerResults
		fmt.Printf("  结果 #%d: %d\n", r, result)
	}

	fmt.Println("\n✅ 所有 goroutine+channel 示例完成")
}

// worker — 工作池中的单个 worker
// jobs   <-chan int  (receive-only): 从 channel 接收任务
// results chan<- int (send-only):    向 channel 发送结果
func worker(id int, jobs <-chan int, results chan<- int) {
	for job := range jobs {
		fmt.Printf("  Worker #%d 开始处理 job %d\n", id, job)
		time.Sleep(20 * time.Millisecond) // 模拟工作
		results <- job * 2                // 返回结果
		fmt.Printf("  Worker #%d 完成 job %d -> %d\n", id, job, job*2)
	}
}