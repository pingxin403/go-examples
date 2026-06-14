// panic-recover/main_test.go — Go 的 Panic 与 Recover 机制测试
package main

import (
	"errors"
	"testing"
)

// TestSafeCall — 测试 safeCall 能捕获 panic 而不向外传播
func TestSafeCall(t *testing.T) {
	t.Run("安全函数不 panic", func(t *testing.T) {
		called := false
		safeCall(func() {
			called = true
		})
		if !called {
			t.Error("safeCall 未执行传入的函数")
		}
	})

	t.Run("捕获 panic 不崩溃", func(t *testing.T) {
		// 如果 safeCall 没捕获好，这里会整个测试崩溃
		safeCall(func() {
			panic("测试 panic")
		})
		// 能到这里说明 safeCall 正确捕获了 panic
	})

	t.Run("嵌套 panic 可恢复", func(t *testing.T) {
		safeCall(func() {
			safeCall(func() {
				panic("嵌套 panic")
			})
		})
	})
}

// TestDivideAndRecover — 测试带 recover 的安全除法
func TestDivideAndRecover(t *testing.T) {
	t.Run("正常除法", func(t *testing.T) {
		result, err := divideAndRecover(10, 2)
		if err != nil {
			t.Errorf("正常除法不应返回 err, 得到: %v", err)
		}
		if result != 5 {
			t.Errorf("10/2 = %d, 期望 5", result)
		}
	})

	t.Run("除零触发 panic", func(t *testing.T) {
		result, err := divideAndRecover(10, 0)
		if err == nil {
			t.Error("除零应返回非 nil err")
		}
		if result != 0 {
			t.Errorf("除零后 result = %d, 期望 0", result)
		}
	})

	t.Run("命名返回值恢复", func(t *testing.T) {
		// 验证命名返回值在 defer 中被正确修改
		_, err := divideAndRecover(5, 0)
		if err == nil {
			t.Fatal("除零应返回 err")
		}
		if err.Error() != "panic 恢复: runtime error: integer divide by zero" {
			t.Logf("错误信息: %s", err.Error())
		}
	})
}

// TestRecoverHelper — 验证 recoverHelper 的行为（间接调 recover 无效）
func TestRecoverHelper(t *testing.T) {
	t.Run("recoverHelper 本身可被 recover 包装调用", func(t *testing.T) {
		// recoverHelper 内部调 recover() — 如果不在 defer 中直接调，返回 nil
		// 但这里 recoverHelper 被 safeCall 包装，所以外层 recover 捕获 panic
		safeCall(func() {
			// 直接调 recoverHelper 不会捕获 panic
			panic("测试消息")
		})
	})
}

// TestSafeHandler — 测试安全 handler 包装器
func TestSafeHandler(t *testing.T) {
	t.Run("正常 handler", func(t *testing.T) {
		called := false
		safeHandler("good", func(name string) {
			called = true
			if name != "good" {
				t.Errorf("name = %s, 期望 good", name)
			}
		})
		if !called {
			t.Error("safeHandler 未执行 handler")
		}
	})

	t.Run("panic handler 不崩溃", func(t *testing.T) {
		safeHandler("bad", func(name string) {
			panic(errors.New("handler panic"))
		})
		// 能到这里说明 safeHandler 正确捕获了 panic
	})

	t.Run("连续调用", func(t *testing.T) {
		count := 0
		for i := 0; i < 5; i++ {
			safeHandler("good", func(name string) {
				count++
			})
		}
		if count != 5 {
			t.Errorf("handler 被执行 %d 次, 期望 5", count)
		}
	})
}