package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("=== Go 控制流综合演示 ===\n")

	// ============================================================
	// 1. if — 支持初始化语句
	// ============================================================
	fmt.Println("--- if 初始化语句 ---")

	// 在 if 中声明变量，作用域仅限于 if-else 块
	if score := 85; score >= 90 {
		fmt.Printf("score=%d: 优秀!\n", score)
	} else if score >= 80 {
		fmt.Printf("score=%d: 良好!\n", score)
	} else if score >= 60 {
		fmt.Printf("score=%d: 及格\n", score)
	} else {
		fmt.Printf("score=%d: 不及格 ❌\n", score)
	}
	// 这里不能访问 score — 编译错误
	// fmt.Println(score) // undefined

	// 常见用法：检查函数返回值
	if n, err := fmt.Println("if 中调用函数并检查返回值"); err != nil {
		fmt.Printf("输出失败: %v\n", err)
	} else {
		fmt.Printf("  成功输出 %d 个字节\n", n)
	}

	// ============================================================
	// 2. for — Go 只有 for，没有 while/do-while
	// ============================================================
	fmt.Println("\n--- for 循环: 经典三部分 ---")

	// 形式一: 经典 for (初始化; 条件; 后置)
	sum := 0
	for i := 1; i <= 5; i++ {
		sum += i
		fmt.Printf("  i=%d, 当前 sum=%d\n", i, sum)
	}
	fmt.Printf("  1+2+3+4+5 = %d\n", sum)

	// 形式二: 类似 while — 只有条件部分
	fmt.Println("\n--- for 循环: while 风格 ---")
	n := 1
	for n < 100 {
		n *= 2
		fmt.Printf("  n=%d\n", n)
	}
	fmt.Printf("  最终 n=%d (超过 100 时退出)\n", n)

	// 形式三: 无限循环 + break
	fmt.Println("\n--- for 无限循环 + break ---")
	count := 0
	for {
		count++
		if count > 3 {
			break
		}
		fmt.Printf("  第 %d 次无限循环\n", count)
	}

	// continue: 跳过本次循环
	fmt.Println("\n--- continue 示例 ---")
	for i := 1; i <= 10; i++ {
		if i%2 != 0 {
			continue // 跳过奇数
		}
		fmt.Printf("  %d 是偶数\n", i)
	}

	// ============================================================
	// 3. for range — 遍历集合
	// ============================================================
	fmt.Println("\n--- for range ---")

	// 3a. 遍历 slice
	fruits := []string{"苹果", "香蕉", "橘子", "葡萄"}
	fmt.Println("遍历 slice:")
	for i, fruit := range fruits {
		fmt.Printf("  fruits[%d] = %q\n", i, fruit)
	}
	// 忽略索引
	fmt.Println("只取值，忽略索引:")
	for _, fruit := range fruits {
		fmt.Printf("  %s\n", fruit)
	}

	// 3b. 遍历 map
	fmt.Println("遍历 map (顺序不固定):")
	scores := map[string]int{"张三": 92, "李四": 78, "王五": 88}
	for name, score := range scores {
		fmt.Printf("  %s: %d 分\n", name, score)
	}
	// 只遍历 key
	fmt.Println("只遍历 key:")
	for name := range scores {
		fmt.Printf("  %s\n", name)
	}

	// 3c. 遍历字符串 (按 rune 遍历)
	fmt.Println("遍历字符串 (按 rune):")
	s := "Hello 世界"
	for i, r := range s {
		fmt.Printf("  s[%d] = %c (U+%04X)\n", i, r, r)
	}

	// 3d. 遍历 channel (先启动 goroutine 发送数据)
	fmt.Println("遍历 channel:")
	ch := make(chan string)
	go func() {
		// 发送者
		messages := []string{"msg1", "msg2", "msg3"}
		for _, m := range messages {
			ch <- m
		}
		close(ch) // 必须关闭，否则 range 死锁
	}()
	for msg := range ch {
		fmt.Printf("  收到: %s\n", msg)
	}

	// ============================================================
	// 4. switch — 支持表达式和 tagless 两种形式
	// ============================================================
	fmt.Println("\n--- switch 表达式 ---")

	// 4a. switch 带表达式
	day := time.Now().Weekday()
	switch day {
	case time.Saturday, time.Sunday:
		fmt.Printf("今天是 %s，周末啦！🎉\n", day)
	default:
		fmt.Printf("今天是 %s，工作日 😅\n", day)
	}

	// switch 中的 case 默认不穿透（不用 break）
	// 如需穿透，使用 fallthrough

	// 4b. switch 带初始化语句
	switch hour := time.Now().Hour(); {
	case hour < 6:
		fmt.Printf("现在 %d 点，深夜了 🌙\n", hour)
	case hour < 12:
		fmt.Printf("现在 %d 点，上午好 ☀️\n", hour)
	case hour < 18:
		fmt.Printf("现在 %d 点，下午好 🌤️\n", hour)
	default:
		fmt.Printf("现在 %d 点，晚上好 🌆\n", hour)
	}

	// 4c. switch tagless — 类似 if-else 链，但更清晰
	fmt.Println("\n--- switch tagless ---")
	score := 75
	switch {
	case score >= 90:
		fmt.Println("优秀")
	case score >= 80:
		fmt.Println("良好")
	case score >= 60:
		fmt.Println("及格")
	default:
		fmt.Println("不及格")
	}

	// 4d. switch 与 fallthrough
	fmt.Println("\n--- switch fallthrough ---")
	switch num := 2; num {
	case 1:
		fmt.Println("  case 1")
		// 没有 fallthrough 就不会执行下面的 case
	case 2:
		fmt.Println("  case 2 — 这里有 fallthrough")
		fallthrough // 强制执行下一个 case
	case 3:
		fmt.Println("  case 3 — 被 fallthrough 穿透到了这里")
	case 4:
		fmt.Println("  case 4 — 正常 case，不会被执行")
	}

	// ============================================================
	// 5. goto — 少用，但某些场景很合适
	// ============================================================
	fmt.Println("\n--- goto ---")
	fmt.Println("开始处理...")

	// 模拟一个需要重试的场景
	retries := 3
	for i := 0; i < retries; i++ {
		result := rand.Intn(10) // 0-9 的随机数
		if result > 7 {
			fmt.Printf("  第 %d 次尝试成功 (result=%d)\n", i+1, result)
			goto success // 跳转到成功标签
		}
		fmt.Printf("  第 %d 次尝试失败 (result=%d)，重试中...\n", i+1, result)
	}
	// 所有重试都失败
	goto failure

success:
	fmt.Println("✅ 操作成功完成!")
	goto cleanup

failure:
	fmt.Println("❌ 操作失败，全部重试用完")

cleanup:
	fmt.Println("🧹 清理资源... 演示结束")
}

// init 在 main 之前自动执行
func init() {
	// 初始化随机种子
	rand.Seed(time.Now().UnixNano())
}