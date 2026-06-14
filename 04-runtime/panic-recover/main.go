// panic-recover/main.go — Go 的 Panic 与 Recover 机制演示
//
// 展示 panic 导致程序崩溃、defer + recover 捕获 panic、
// 命名返回值恢复、goroutine 中 panic 不可恢复、
// 多级 defer 的执行顺序、以及实用的安全包装器。
//
// 注意：每个演示都包在独立的函数中，这样 defer+recover
// 不会影响 main 的后续流程（defer 只在包裹函数作用域内生效）。
package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"
)

func main() {
	// 每个演示用独立函数包装，确保 defer 作用域不扩散到 main
	section1()
	section2()
	section3()
	section4MultiLevelDefer()
	section5GoroutinePanic()
	section6DirectRecoverOnly()
	section7SafeHandler()
	section8StackTrace()

	fmt.Println("\nPanic 与 Recover 概念演示完毕 ✓")
}

// ============================================================
// 1. 基本 panic：未捕获的 panic 导致程序崩溃
// ============================================================
func section1() {
	fmt.Println("=== 1. 基本 panic（注释掉以避免崩溃） ===")
	fmt.Println("   // panic(\"something went wrong\")  // 这行会崩溃程序")
	fmt.Println("   未捕获的 panic → 栈打印 → os.Exit(非0)")
}

// ============================================================
// 2. defer + recover：像 try-catch 一样捕获 panic
// ============================================================
func section2() {
	fmt.Println("\n=== 2. defer + recover 捕获 panic ===")

	safeCall(func() {
		fmt.Println("  在 safeCall 内部执行...")
		panic("出错了！")
	})
	fmt.Println("  程序继续正常执行 ✓")
}

// ============================================================
// 3. 命名返回值 + recover
// ============================================================
func section3() {
	fmt.Println("\n=== 3. 命名返回值 + recover ===")

	result, err := divideAndRecover(10, 0)
	fmt.Printf("  10 / 0 = %d, err = %v\n", result, err)

	result, err = divideAndRecover(10, 2)
	fmt.Printf("  10 / 2 = %d, err = %v\n", result, err)
}

// ============================================================
// 4. 多级 defer 与 panic — defer 是栈（LIFO）顺序
// ============================================================
func section4MultiLevelDefer() {
	fmt.Println("\n=== 4. 多级 defer 执行顺序 ===")

	// 内部函数 + defer 包装，模拟 nested defer 效果
	func() {
		defer fmt.Println("  外层 defer 2 (先注册)")
		defer func() {
			fmt.Println("  外层 defer 1 中的 recover:")
			if r := recover(); r != nil {
				fmt.Printf("    recover 捕获: %v\n", r)
			}
		}()

		// 内层作用域
		func() {
			defer fmt.Println("  内层 defer 3 (后注册)")
			defer fmt.Println("  内层 defer 2")
			defer fmt.Println("  内层 defer 1 (最先生效)")
			fmt.Println("  开始触发 panic...")
			panic("panic in inner scope")
		}()
	}()

	fmt.Println("  recover 后继续执行 ✓ (所有 defer 在闭包作用域内)")
}

// ============================================================
// 5. panic 在 goroutine 中 — 不能被父 goroutine 恢复
// ============================================================
func section5GoroutinePanic() {
	fmt.Println("\n=== 5. Goroutine 中的 panic 不可从外部恢复 ===")

	// 错误示范（注释掉，否则会崩溃）
	/*
		go func() {
			panic("goroutine panic") // 整个进程崩溃
		}()
		time.Sleep(time.Second)
		recover() // 没用！不同 goroutine
	*/

	// 正确做法：在 goroutine 内部 recover
	fmt.Println("  Goroutine 中的 panic 必须在 goroutine 内部 recover:")

	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("    goroutine 内部 recover: %v\n", r)
			}
			done <- true
		}()
		panic("goroutine 内部 panic")
	}()
	<-done
	fmt.Println("  主 goroutine 不受影响 ✓")
}

// ============================================================
// 6. recover 只在 defer 函数中直接调用有效
// ============================================================
func section6DirectRecoverOnly() {
	fmt.Println("\n=== 6. recover 只在 defer 中直接调才有效 ===")

	// 错误示范：recover 不在 defer 中直接调用
	func() {
		defer func() {
			// 间接调用 recover — recoverHelper 中的 recover() 返回 nil，
			// 因为 recover 必须在 defer 函数中直接调用
			recoverHelper()
		}()
		// 这个 panic 不会被捕获！程序会崩溃…
		// 我们用 safeCall 包装以免影响后续演示
	}()

	// 用 safeCall 演示间接 recover 失败
	safeCall(func() {
		func() {
			defer func() {
				recoverHelper() // recover 不在此 defer 中直接调用
			}()
			panic("通过 recoverHelper 间接调用 recover → 无法捕获!")
		}()
	})
	fmt.Println("  上面 panic 通过 recoverHelper 间接调用 → 逃逸到 safeCall 被捕获 ✓")

	// 正确模式：recover 必须在 defer 函数中直接调用
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("    正确模式: recover 捕获到 %v\n", r)
				fmt.Println("    recover 必须直接在 defer 函数中调用才有效")
			}
		}()
		panic("演示 recover 的正确用法")
	}()
}

// ============================================================
// 7. 实用模式：安全 HTTP Handler 包装器
// ============================================================
func section7SafeHandler() {
	fmt.Println("\n=== 7. 实用模式：安全 Handler 包装器 ===")

	// 模拟一个可能 panic 的处理器
	handler := func(name string) {
		fmt.Printf("  处理请求: %s\n", name)
		if name == "bad" {
			panic(errors.New("发生了未预期的错误"))
		}
		fmt.Printf("  请求 %s 正常完成 ✓\n", name)
	}

	// 用安全包装器执行
	for _, name := range []string{"good", "bad", "good2"} {
		safeHandler(name, handler)
	}
}

// ============================================================
// 8. recover 后获取栈信息
// ============================================================
func section8StackTrace() {
	fmt.Println("\n=== 8. recover 后获取栈信息 ===")

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  捕获到 panic: %v\n", r)
				// 获取栈信息（调试用）
				buf := make([]byte, 1024)
				n := runtime.Stack(buf, false)
				fmt.Printf("  当前栈信息:\n%s\n", buf[:n])
			}
		}()
		panic("stack trace demo")
	}()
}

// ============================================================
// 辅助函数
// ============================================================

// safeCall — 通用安全调用包装，保护外部不因 panic 崩溃
func safeCall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  safeCall 捕获 panic: %v\n", r)
		}
	}()
	fn()
}

// divideAndRecover — 安全的除法，使用命名返回值 + recover
func divideAndRecover(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic 恢复: %v", r)
			result = 0 // 命名返回值可在 defer 中修改
		}
	}()
	result = a / b // 如果 b=0，触发 panic
	return
}

// recoverHelper — 演示 recover 间接调用无效
func recoverHelper() {
	if r := recover(); r != nil {
		fmt.Printf("  recoverHelper 捕获: %v\n", r)
	}
}

// safeHandler — 类似 HTTP handler 的安全包装
func safeHandler(name string, handler func(string)) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[安全包装] handler %s panic: %v, 已恢复\n", name, r)
		}
	}()
	handler(name)
}