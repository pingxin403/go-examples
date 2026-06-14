package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ------------------------------------------------------------
// 1. 结构体定义与初始化
// ------------------------------------------------------------

// Person 基础结构体
type Person struct {
	Name string
	Age  int
}

// demoInit 演示结构体初始化的几种方式
func demoInit() {
	fmt.Println("=== 结构体初始化 ===")

	// 方式 1: 零值初始化
	var p1 Person
	fmt.Printf("零值: %+v\n", p1)

	// 方式 2: 字面量（按字段顺序，不推荐）
	p2 := Person{"Alice", 30}
	fmt.Printf("顺序字面量: %+v\n", p2)

	// 方式 3: 字面量（按字段名，推荐）
	p3 := Person{Name: "Bob", Age: 25}
	fmt.Printf("命名字面量: %+v\n", p3)

	// 方式 4: 部分初始化（未指定的字段为零值）
	p4 := Person{Name: "Charlie"}
	fmt.Printf("部分初始化: %+v\n", p4)

	// 方式 5: new 关键字（返回指针）
	p5 := new(Person)
	p5.Name = "Dave"
	fmt.Printf("new 指针: %+v\n", p5)
}

// ------------------------------------------------------------
// 2. 匿名结构体
// ------------------------------------------------------------

// demoAnonStruct 演示匿名结构体
func demoAnonStruct() {
	fmt.Println("\n=== 匿名结构体 ===")

	// 一次性使用的结构体，无需提前定义
	book := struct {
		Title  string
		Author string
		Pages  int
	}{
		Title:  "Go 程序设计语言",
		Author: "Alan A. A. Donovan",
		Pages:  400,
	}
	fmt.Printf("匿名结构体: %+v\n", book)

	// 常用于 JSON 解析中的临时结构
	_ = struct {
		Status int    `json:"status"`
		Msg    string `json:"message"`
	}{Status: 200, Msg: "OK"}
}

// ------------------------------------------------------------
// 3. 结构体嵌入（组合优于继承）
// ------------------------------------------------------------

// Address 可复用的地址结构
type Address struct {
	City    string
	Street  string
	ZipCode string
}

// Employee 通过嵌入 Address 实现组合
type Employee struct {
	Name    string
	Age     int
	Address            // 匿名字段嵌入（提升）
	Phone   string     // 具名字段
	Manager *Employee  // 自引用指针
}

// demoEmbedding 演示结构体嵌入
func demoEmbedding() {
	fmt.Println("\n=== 结构体嵌入 ===")

	e := Employee{
		Name: "Alice",
		Age:  30,
		Address: Address{
			City:    "北京",
			Street:  "中关村大街",
			ZipCode: "100000",
		},
		Phone: "138-0000-0000",
	}

	// 可以直接访问被提升的字段
	fmt.Printf("Name: %s\n", e.Name)
	fmt.Printf("City: %s（直接访问嵌入字段）\n", e.City)
	fmt.Printf("Street: %s（也可以通过 Address 访问）\n", e.Address.Street)

	// 嵌入结构体的方法也会被提升（见下面的 String() 演示）
}

// ------------------------------------------------------------
// 4. Struct Tags 与 JSON 序列化
// ------------------------------------------------------------

// Config 演示 struct tags 的多种用法
type Config struct {
	Host    string `json:"host"`               // JSON 字段名为 host
	Port    int    `json:"port,omitempty"`       // 零值时忽略
	Debug   bool   `json:"debug,omitempty"`      // false 时忽略
	Secret  string `json:"-"`                    // 永远不参与序列化
	Timeout int    `json:"timeout_seconds"`      // 自定义字段名
	Tags    []string `json:"tags,omitempty"`     // 空 slice 时忽略
}

// User 演示常见的结构体 tag 用法
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
}

// demoJSON 演示 JSON 序列化/反序列化
func demoJSON() {
	fmt.Println("\n=== JSON 序列化 ===")

	// 序列化
	cfg := Config{
		Host:    "localhost",
		Port:    8080,
		Debug:   false,
		Secret:  "my-secret",
		Timeout: 30,
		Tags:    []string{"dev", "test"},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Config JSON:\n%s\n", string(data))

	// 验证：Secret 不会出现，Debug 因为是 false + omitempty 被忽略
	fmt.Println("注意: secret 已隐藏, debug=false 已省略")

	// 反序列化
	jsonStr := `{"id":1,"name":"Alice","email":"alice@example.com","created_at":"2024-01-01T00:00:00Z","is_admin":true}`
	var user User
	if err := json.Unmarshal([]byte(jsonStr), &user); err != nil {
		panic(err)
	}
	fmt.Printf("反序列化 User: %+v\n", user)

	// omitempty 效果：零值字段被忽略
	minimal := Config{Host: "localhost"}
	minJSON, _ := json.Marshal(minimal)
	fmt.Printf("最小配置 JSON: %s\n", minJSON)
}

// ------------------------------------------------------------
// 5. 值接收者 vs 指针接收者
// ------------------------------------------------------------

// Rectangle 矩形结构体
type Rectangle struct {
	Width  float64
	Height float64
}

// Area 值接收者（不修改原对象）
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Scale 指针接收者（修改原对象）
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// String 实现 fmt.Stringer 接口（值接收者也可以）
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.1f x %.1f, area=%.1f)", r.Width, r.Height, r.Area())
}

// demoReceiver 演示值接收者和指针接收者的区别
func demoReceiver() {
	fmt.Println("\n=== 值接收者 vs 指针接收者 ===")

	r := Rectangle{Width: 10, Height: 5}

	// 值接收者方法：不会修改原对象
	fmt.Println(r.String())
	fmt.Printf("调用 Area(): %.1f\n", r.Area())

	// 指针接收者方法：会修改原对象
	r.Scale(2)
	fmt.Printf("调用 Scale(2) 后: %s\n", r.String())

	// 重要：即使通过值调用指针方法，Go 会自动取地址
	r2 := Rectangle{Width: 3, Height: 4}
	r2.Scale(1.5) // 等价于 (&r2).Scale(1.5)
	fmt.Printf("r2: %s\n", r2.String())
}

// ------------------------------------------------------------
// 6. 实战：User 结构体与校验
// ------------------------------------------------------------

// UserInfo 用户信息结构体
type UserInfo struct {
	Username string
	Email    string
	Age      int
}

// Validate 校验用户信息的合法性
func (u UserInfo) Validate() error {
	if len(u.Username) < 3 || len(u.Username) > 20 {
		return errors.New("用户名长度必须在 3-20 之间")
	}
	if !stringsContains(u.Email, "@") {
		return errors.New("邮箱格式无效")
	}
	if u.Age < 0 || u.Age > 150 {
		return errors.New("年龄必须在 0-150 之间")
	}
	return nil
}

// Greet 返回问候语（值接收者，只读）
func (u UserInfo) Greet() string {
	return fmt.Sprintf("你好！我是 %s（%d 岁）", u.Username, u.Age)
}

// SetAge 修改年龄（指针接收者，需要修改）
func (u *UserInfo) SetAge(age int) error {
	if age < 0 || age > 150 {
		return errors.New("年龄必须在 0-150 之间")
	}
	u.Age = age
	return nil
}

// 辅助函数
func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// demoUserValidation 演示实际的结构体校验场景
func demoUserValidation() {
	fmt.Println("\n=== 实战：User 结构体与校验 ===")

	users := []UserInfo{
		{Username: "alice", Email: "alice@example.com", Age: 28},
		{Username: "ab", Email: "invalid", Age: -1},
		{Username: "this-username-is-way-too-long-and-should-fail", Email: "test@test.com", Age: 20},
	}

	for i, u := range users {
		fmt.Printf("\n用户 %d:\n", i+1)
		if err := u.Validate(); err != nil {
			fmt.Printf("  校验失败: %v\n", err)
			continue
		}
		fmt.Printf("  %s\n", u.Greet())

		// 指针接收者演示
		newAge := u.Age + 1
		if err := u.SetAge(newAge); err != nil {
			fmt.Printf("  设置年龄失败: %v\n", err)
		} else {
			fmt.Printf("  生日+1: 现在 %d 岁\n", u.Age)
		}
	}
}

// NewPerson 创建并返回一个 Person 指针（工厂函数）
func NewPerson(name string, age int) *Person {
	return &Person{Name: name, Age: age}
}

// NewRectangle 创建并返回一个 Rectangle（工厂函数）
func NewRectangle(width, height float64) Rectangle {
	return Rectangle{Width: width, Height: height}
}

// ------------------------------------------------------------
// main
// ------------------------------------------------------------

func main() {
	fmt.Println("========================================")
	fmt.Println("  02-datastructures / struct")
	fmt.Println("========================================")

	demoInit()
	demoAnonStruct()
	demoEmbedding()
	demoJSON()
	demoReceiver()
	demoUserValidation()

	fmt.Println("\n=== 完成 ===")
}