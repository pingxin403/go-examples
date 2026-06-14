package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ============================================================
// select + timer 示例
// 涵盖：select 多路复用、default 非阻塞、Ticker 定时任务、
// Timer 一次性超时、select 超时模式、fan-in 合并
// ============================================================

func main() {
	// ============================================================
	// 1. select — 多 channel 等待（谁先到处理谁）
	// ============================================================
	fmt.Println("=== 1. select 多路复用 ===")

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(40 * time.Millisecond)
		ch1 <- "来自 ch1"
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		ch2 <- "来自 ch2"
	}()

	// 同时等待两个 channel，哪个先到处理哪个
	select {
	case msg := <-ch1:
		fmt.Printf("  收到 ch1: %s\n", msg)
	case msg := <-ch2:
		fmt.Printf("  收到 ch2: %s（先到！）\n", msg)
	}

	// ============================================================
	// 2. select + default — 非阻塞操作
	// ============================================================
	fmt.Println("\n=== 2. select + default 非阻塞 ===")

	nonBlockingCh := make(chan int, 1)

	// 尝试发送（channel 为空，有缓冲，可以发送）
	select {
	case nonBlockingCh <- 42:
		fmt.Println("  发送成功: 42 写入 channel")
	default:
		fmt.Println("  channel 已满，无法发送")
	}

	// 尝试发送（channel 已满）
	select {
	case nonBlockingCh <- 99:
		fmt.Println("  发送成功: 99 写入 channel")
	default:
		fmt.Println("  channel 已满，无法发送 99")
	}

	// 非阻塞读取
	select {
	case val := <-nonBlockingCh:
		fmt.Printf("  读取成功: %d\n", val)
	default:
		fmt.Println("  无数据可读")
	}

	// ============================================================
	// 3. time.Ticker — 周期性任务
	// ============================================================
	fmt.Println("\n=== 3. time.Ticker 定时任务 ===")

	ticker := time.NewTicker(40 * time.Millisecond)
	done := make(chan bool)

	// 启动一个 goroutine 读取 ticker
	go func() {
		time.Sleep(130 * time.Millisecond) // 跑 3 个 tick 左右
		done <- true
	}()

	// 用 for + select 驱动周期性任务
	runs := 0
	for {
		select {
		case t := <-ticker.C:
			runs++
			fmt.Printf("  Tick #%d at %v\n", runs, t.Format("15:04:05.000"))
		case <-done:
			ticker.Stop() // 停止 ticker，释放资源
			fmt.Println("  Ticker 已停止")
		}
		if runs >= 3 {
			ticker.Stop()
			break
		}
	}

	// ============================================================
	// 4. time.Timer — 一次性定时器
	// ============================================================
	fmt.Println("\n=== 4. time.Timer 一次性定时器 ===")

	// Timer 在指定时间后向 C channel 发送一次事件
	timer := time.NewTimer(50 * time.Millisecond)
	fmt.Println("  定时器启动，等待 50ms...")

	<-timer.C
	fmt.Println("  定时器触发 ✅")

	// Timer 可以提前停止
	timer2 := time.NewTimer(1 * time.Hour)
	go func() {
		time.Sleep(10 * time.Millisecond)
		timer2.Stop()
	}()
	time.Sleep(20 * time.Millisecond)
	fmt.Println("  timer2 已停止（不会触发）")

	// Timer 可以复用（Reset）
	timer3 := time.NewTimer(30 * time.Millisecond)
	<-timer3.C
	fmt.Println("  timer3 第一次触发")
	timer3.Reset(30 * time.Millisecond)
	<-timer3.C
	fmt.Println("  timer3 第二次触发 ✅")

	// ============================================================
	// 5. select 超时模式 — 防止无限阻塞
	// ============================================================
	fmt.Println("\n=== 5. select 超时模式 ===")

	slowCh := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond) // 很慢
		slowCh <- "终于完成了"
	}()

	select {
	case result := <-slowCh:
		fmt.Printf("  收到结果: %s\n", result)
	case <-time.After(50 * time.Millisecond): // 超时 50ms
		fmt.Println("  ⏰ 超时！任务太慢了，不等了")
	}

	// ============================================================
	// 6. Fan-in 模式 — 将多个 channel 合并为一个
	// ============================================================
	fmt.Println("\n=== 6. Fan-in — 合并多个 channel ===")

	// fanIn 函数将两个 channel 合并为一个输出 channel
	fanIn := func(inputs ...<-chan int) <-chan int {
		out := make(chan int)

		for _, ch := range inputs {
			// 为每个输入 channel 启动一个 goroutine
			go func(c <-chan int) {
				for v := range c {
					out <- v
				}
			}(ch)
		}

		return out
	}

	// 创建两个数据源
	source1 := make(chan int)
	source2 := make(chan int)

	// 启动生产者
	go func() {
		for i := 1; i <= 3; i++ {
			source1 <- i * 10
			time.Sleep(time.Duration(rand.Intn(20)+10) * time.Millisecond)
		}
		close(source1)
	}()
	go func() {
		for i := 1; i <= 3; i++ {
			source2 <- i*10 + 5
			time.Sleep(time.Duration(rand.Intn(20)+10) * time.Millisecond)
		}
		close(source2)
	}()

	// 合并并消费
	merged := fanIn(source1, source2)
	received := 0
	for v := range merged {
		fmt.Printf("  Fan-in 收到: %d\n", v)
		received++
		if received >= 6 {
			break
		}
	}

	fmt.Println("\n✅ 所有 select+timer 示例完成")
}