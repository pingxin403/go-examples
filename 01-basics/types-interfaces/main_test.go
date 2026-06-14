// main_test.go — 对 types-interfaces 包中接口定义、类型断言、接口嵌入等特性的完整测试套件
//
// 本文件演示 Go 测试的多种模式：
//   - 表格驱动测试（Table-Driven Tests）
//   - 子测试（t.Run）
//   - 接口多态验证测试
//   - 类型断言安全/非安全测试
//   - Stringer 接口实现验证
//   - 接口 nil 陷阱测试
package main

import (
	"testing"
)

// ============================================================
// 1. 接口隐式实现测试（Duck Typing）
// ============================================================

// TestSpeakerInterface 验证 Dog、Cat、Robot 都实现了 Speaker 接口
func TestSpeakerInterface(t *testing.T) {
	tests := []struct {
		name     string
		speaker  Speaker
	}{
		{name: "Dog 实现 Speaker", speaker: Dog{Name: "旺财"}},
		{name: "Cat 实现 Speaker", speaker: Cat{Name: "咪咪"}},
		{name: "Robot 实现 Speaker", speaker: Robot{Model: "R2-D2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证可以调用 Speak 方法而不 panic
			s := tt.speaker.Speak()
			if s == "" {
				t.Errorf("Speak() 返回空字符串")
			}
		})
	}
}

// TestMakeSpeak 验证 makeSpeak 函数接受不同 Speaker 类型
func TestMakeSpeak(t *testing.T) {
	// 验证接受 Speaker 变参不 panic
	makeSpeak(
		Dog{Name: "旺财"},
		Cat{Name: "咪咪"},
		Robot{Model: "R2-D2"},
	)
}

// TestSpeakerSlice 验证 Speaker slice 的多态行为
func TestSpeakerSlice(t *testing.T) {
	speakers := []Speaker{
		Dog{Name: "旺财"},
		Cat{Name: "咪咪"},
		Robot{Model: "R2-D2"},
	}

	if len(speakers) != 3 {
		t.Fatalf("Speaker slice 长度 = %d; 期望 3", len(speakers))
	}
}

// ============================================================
// 2. 空接口 / any 测试
// ============================================================

// TestDescribeAny 验证 describeAny 不 panic
func TestDescribeAny(t *testing.T) {
	// 各种类型都不应 panic
	describeAny(42)
	describeAny("hello")
	describeAny(3.14)
	describeAny(true)
	describeAny(Dog{Name: "小白"})
	describeAny([]int{1, 2, 3})
}

// ============================================================
// 3. 类型断言测试
// ============================================================

// TestSafeTypeAssert 表格驱动测试安全类型断言
func TestSafeTypeAssert(t *testing.T) {
	t.Run("string 断言成功", func(t *testing.T) {
		var v any = "安全字符串"
		s, ok := v.(string)
		if !ok {
			t.Fatal("类型断言应成功")
		}
		if s != "安全字符串" {
			t.Errorf("断言值 = %q; 期望 %q", s, "安全字符串")
		}
	})

	t.Run("int 断言为 string 失败", func(t *testing.T) {
		var v any = 42
		_, ok := v.(string)
		if ok {
			t.Fatal("int 断言为 string 应失败")
		}
	})
}

// TestUnsafeTypeAssert 验证非安全类型断言在错误类型时会 panic
func TestUnsafeTypeAssert(t *testing.T) {
	t.Run("正确类型不 panic", func(t *testing.T) {
		// 不应 panic
		unsafeTypeAssert("hello")
	})
}

// ============================================================
// 4. 类型开关（Type Switch）测试
// ============================================================

// TestTypeSwitch 表格驱动测试类型开关
func TestTypeSwitch(t *testing.T) {
	tests := []struct {
		name string
		val  any
	}{
		{name: "int 类型", val: 42},
		{name: "float64 类型", val: 3.14},
		{name: "string 类型", val: "hello"},
		{name: "bool 类型", val: true},
		{name: "nil 类型", val: nil},
		{name: "Speaker 类型", val: Dog{Name: "旺财"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证 typeSwitch 不 panic
			typeSwitch(tt.val)
		})
	}
}

// ============================================================
// 5. 接口嵌入测试
// ============================================================

// TestSimulatedFile 验证 SimulatedFile 实现了 ReadWriter 接口
func TestSimulatedFile(t *testing.T) {
	file := SimulatedFile{Name: "test.txt"}

	// 验证 Read 方法
	n, err := file.Read([]byte("data"))
	if err != nil {
		t.Fatalf("Read 错误: %v", err)
	}
	if n <= 0 {
		t.Errorf("Read 返回 %d 字节; 期望 > 0", n)
	}

	// 验证 Write 方法
	n, err = file.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write 错误: %v", err)
	}
	if n != 5 {
		t.Errorf("Write 返回 %d; 期望 5", n)
	}

	// 验证 Close 方法
	err = file.Close()
	if err != nil {
		t.Fatalf("Close 错误: %v", err)
	}
}

// TestProcessFile 验证 processFile 接受 ReadWriter 接口
func TestProcessFile(t *testing.T) {
	// 验证不 panic
	file := SimulatedFile{Name: "data.txt"}
	processFile(file)
}

// ============================================================
// 6. Stringer 接口测试
// ============================================================

// TestPersonString 验证 Person 的 String() 方法
func TestPersonString(t *testing.T) {
	tests := []struct {
		name     string
		person   Person
		want     string
	}{
		{name: "正常 Person", person: Person{FirstName: "小明", LastName: "张", Age: 28}, want: "张 小明 (28 岁)"},
		{name: "零值 Person", person: Person{}, want: "  (0 岁)"},
		{name: "仅名字", person: Person{FirstName: "三", LastName: "李"}, want: "李 三 (0 岁)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.person.String()
			if got != tt.want {
				t.Errorf("Person.String() = %q; 期望 %q", got, tt.want)
			}
		})
	}
}

// TestIPAddrString 验证 IPAddr 的 String() 方法
func TestIPAddrString(t *testing.T) {
	tests := []struct {
		name string
		ip   IPAddr
		want string
	}{
		{name: "localhost", ip: IPAddr{127, 0, 0, 1}, want: "127.0.0.1"},
		{name: "全零地址", ip: IPAddr{0, 0, 0, 0}, want: "0.0.0.0"},
		{name: "DNS 地址", ip: IPAddr{8, 8, 8, 8}, want: "8.8.8.8"},
		{name: "广播地址", ip: IPAddr{255, 255, 255, 255}, want: "255.255.255.255"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ip.String()
			if got != tt.want {
				t.Errorf("IPAddr.String() = %q; 期望 %q", got, tt.want)
			}
		})
	}
}

// TestMoneyString 验证 Money 的 String() 方法
func TestMoneyString(t *testing.T) {
	tests := []struct {
		name string
		money Money
		want string
	}{
		{name: "人民币", money: Money{Amount: 99.99, Currency: "¥"}, want: "¥99.99"},
		{name: "美元", money: Money{Amount: 19.99, Currency: "$"}, want: "$19.99"},
		{name: "零元", money: Money{Amount: 0, Currency: "¥"}, want: "¥0.00"},
		{name: "欧元", money: Money{Amount: 1234.56, Currency: "€"}, want: "€1234.56"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.money.String()
			if got != tt.want {
				t.Errorf("Money.String() = %q; 期望 %q", got, tt.want)
			}
		})
	}
}

// ============================================================
// 7. 接口 nil 值陷阱测试
// ============================================================

// TestInterfaceNilTraps 验证接口值的 nil 行为
func TestInterfaceNilTraps(t *testing.T) {
	t.Run("未初始化的接口值为 nil", func(t *testing.T) {
		var s Speaker
		if s != nil {
			t.Error("未初始化的 Speaker 接口应为 nil")
		}
	})

	t.Run("赋值后接口不为 nil", func(t *testing.T) {
		var s Speaker = Dog{Name: "旺财"}
		if s == nil {
			t.Error("赋值后的 Speaker 接口不应为 nil")
		}
	})

	t.Run("包含 nil 指针的接口 != nil", func(t *testing.T) {
		var p *Dog
		var s Speaker = p
		if s == nil {
			// 这是 Go 的一个常见陷阱
			// 接口值为 nil 当且仅当 type 和 value 都是 nil
			// 当接口持有 *Dog(nil) 时，type 不为 nil，所以接口不为 nil
			t.Error("包含 nil 指针的接口不应为 nil（常见陷阱）")
		}
	})
}

// ============================================================
// 8. 类型转换测试
// ============================================================

// TestTypeConversionString 验证 rune 到 string 的类型转换
func TestTypeConversionString(t *testing.T) {
	s := string(rune(65))
	if s != "A" {
		t.Errorf("string(rune(65)) = %q; 期望 'A'", s)
	}
}

// TestDogSpeakOutput 验证 Dog 的具体输出
func TestDogSpeakOutput(t *testing.T) {
	tests := []struct {
		name string
		dog  Dog
		want string
	}{
		{name: "旺财", dog: Dog{Name: "旺财"}, want: "汪汪！我是 旺财 🐕"},
		{name: "小白", dog: Dog{Name: "小白"}, want: "汪汪！我是 小白 🐕"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dog.Speak()
			if got != tt.want {
				t.Errorf("Dog.Speak() = %q; 期望 %q", got, tt.want)
			}
		})
	}
}

// TestCatSpeakOutput 验证 Cat 的具体输出
func TestCatSpeakOutput(t *testing.T) {
	tests := []struct {
		name string
		cat  Cat
		want string
	}{
		{name: "咪咪", cat: Cat{Name: "咪咪"}, want: "喵～我是 咪咪 🐱"},
		{name: "花花", cat: Cat{Name: "花花"}, want: "喵～我是 花花 🐱"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cat.Speak()
			if got != tt.want {
				t.Errorf("Cat.Speak() = %q; 期望 %q", got, tt.want)
			}
		})
	}
}