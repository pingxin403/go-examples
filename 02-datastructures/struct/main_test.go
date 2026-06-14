// main_test.go — 对 struct 包的核心操作进行表格驱动测试
//
// 覆盖范围：
//   - 结构体初始化、工厂函数、嵌入
//   - JSON 序列化/反序列化与 struct tags
//   - 值接收者 vs 指针接收者方法
//   - 结构体校验与业务方法
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================
// 1. 结构体初始化
// ============================================================

// TestNewPerson 测试 Person 工厂函数
func TestNewPerson(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		age      int
		wantName string
		wantAge  int
	}{
		{name: "正常参数", userName: "Alice", age: 30, wantName: "Alice", wantAge: 30},
		{name: "零岁", userName: "Baby", age: 0, wantName: "Baby", wantAge: 0},
		{name: "空名字", userName: "", age: 25, wantName: "", wantAge: 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPerson(tt.userName, tt.age)
			if p.Name != tt.wantName {
				t.Errorf("Name = %q, 期望 %q", p.Name, tt.wantName)
			}
			if p.Age != tt.wantAge {
				t.Errorf("Age = %d, 期望 %d", p.Age, tt.wantAge)
			}
		})
	}
}

// TestNewRectangle 测试 Rectangle 工厂函数
func TestNewRectangle(t *testing.T) {
	tests := []struct {
		name           string
		width, height  float64
		wantWidth      float64
		wantHeight     float64
	}{
		{name: "正常矩形", width: 10, height: 5, wantWidth: 10, wantHeight: 5},
		{name: "正方形", width: 4, height: 4, wantWidth: 4, wantHeight: 4},
		{name: "零值", width: 0, height: 0, wantWidth: 0, wantHeight: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRectangle(tt.width, tt.height)
			if r.Width != tt.wantWidth {
				t.Errorf("Width = %.1f, 期望 %.1f", r.Width, tt.wantWidth)
			}
			if r.Height != tt.wantHeight {
				t.Errorf("Height = %.1f, 期望 %.1f", r.Height, tt.wantHeight)
			}
		})
	}
}

// ============================================================
// 2. Struct 嵌入
// ============================================================

// TestEmbedding 测试结构体嵌入与字段提升
func TestEmbedding(t *testing.T) {
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

	t.Run("直接访问嵌入字段", func(t *testing.T) {
		if e.City != "北京" {
			t.Errorf("e.City = %q, 期望 北京", e.City)
		}
	})

	t.Run("通过类型名访问嵌入字段", func(t *testing.T) {
		if e.Address.Street != "中关村大街" {
			t.Errorf("e.Address.Street = %q, 期望 中关村大街", e.Address.Street)
		}
	})

	t.Run("具名字段正常访问", func(t *testing.T) {
		if e.Phone != "138-0000-0000" {
			t.Errorf("e.Phone = %q, 期望 138-0000-0000", e.Phone)
		}
	})

	t.Run("嵌入的零值字段可访问", func(t *testing.T) {
		if e.Manager != nil {
			t.Errorf("e.Manager = %v, 期望 nil", e.Manager)
		}
	})
}

// ============================================================
// 3. JSON 序列化/反序列化
// ============================================================

// TestConfigJSON 测试 Config 的 JSON 序列化
func TestConfigJSON(t *testing.T) {
	t.Run("完整配置序列化", func(t *testing.T) {
		cfg := Config{
			Host:    "localhost",
			Port:    8080,
			Debug:   false,
			Secret:  "my-secret",
			Timeout: 30,
			Tags:    []string{"dev", "test"},
		}
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("json.Marshal 失败: %v", err)
		}

		jsonStr := string(data)

		// host 字段应出现
		if !strings.Contains(jsonStr, `"host"`) {
			t.Error("host 字段应出现在 JSON 中")
		}

		// Secret 有 `json:"-"`，不应出现
		if strings.Contains(jsonStr, "my-secret") || strings.Contains(jsonStr, "secret") {
			t.Error("Secret 字段不应出现在 JSON 中（标签为 json:\"-\"）")
		}

		// Debug 为 false + omitempty，应被忽略
		if strings.Contains(jsonStr, "debug") {
			t.Error("Debug=false 且 omitempty，应被省略")
		}

		// timeout_seconds 是自定义字段名
		if !strings.Contains(jsonStr, "timeout_seconds") {
			t.Error("应使用自定义字段名 timeout_seconds")
		}
	})

	t.Run("最小配置序列化", func(t *testing.T) {
		minimal := Config{Host: "localhost"}
		data, _ := json.Marshal(minimal)
		jsonStr := string(data)

		if !strings.Contains(jsonStr, `"host"`) {
			t.Error("host 字段应出现")
		}
	})
}

// TestUserJSON 测试 User 的 JSON 序列化与反序列化
func TestUserJSON(t *testing.T) {
	t.Run("反序列化", func(t *testing.T) {
		jsonStr := `{"id":1,"name":"Alice","email":"alice@example.com","created_at":"2024-01-01T00:00:00Z","is_admin":true}`
		var user User
		if err := json.Unmarshal([]byte(jsonStr), &user); err != nil {
			t.Fatalf("json.Unmarshal 失败: %v", err)
		}
		if user.ID != 1 {
			t.Errorf("ID = %d, 期望 1", user.ID)
		}
		if user.Name != "Alice" {
			t.Errorf("Name = %q, 期望 Alice", user.Name)
		}
		if !user.IsAdmin {
			t.Error("IsAdmin 应为 true")
		}
	})

	t.Run("序列化使用的字段名来自 json tag", func(t *testing.T) {
		u := User{ID: 42, Name: "Bob", Email: "bob@test.com", IsAdmin: false}
		data, err := json.Marshal(u)
		if err != nil {
			t.Fatalf("json.Marshal 失败: %v", err)
		}
		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"created_at"`) {
			t.Error("字段名应为 created_at（来自 json tag）")
		}
	})
}

// ============================================================
// 4. 值接收者 vs 指针接收者
// ============================================================

// TestRectangle_Area 测试 Rectangle.Area 值接收者方法
func TestRectangle_Area(t *testing.T) {
	tests := []struct {
		name           string
		width, height  float64
		want           float64
	}{
		{name: "标准矩形", width: 10, height: 5, want: 50},
		{name: "正方形", width: 4, height: 4, want: 16},
		{name: "零宽", width: 0, height: 5, want: 0},
		{name: "小数", width: 3.5, height: 2.0, want: 7.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rectangle{Width: tt.width, Height: tt.height}
			got := r.Area()
			if got != tt.want {
				t.Errorf("Area() = %.1f, 期望 %.1f", got, tt.want)
			}
		})
	}
}

// TestRectangle_Scale 测试 Rectangle.Scale 指针接收者修改原对象
func TestRectangle_Scale(t *testing.T) {
	tests := []struct {
		name           string
		width, height  float64
		factor         float64
		wantWidth      float64
		wantHeight     float64
	}{
		{name: "2倍缩放", width: 10, height: 5, factor: 2, wantWidth: 20, wantHeight: 10},
		{name: "缩小一半", width: 10, height: 5, factor: 0.5, wantWidth: 5, wantHeight: 2.5},
		{name: "1倍（不变）", width: 7, height: 3, factor: 1, wantWidth: 7, wantHeight: 3},
		{name: "0倍为零", width: 10, height: 5, factor: 0, wantWidth: 0, wantHeight: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rectangle{Width: tt.width, Height: tt.height}
			r.Scale(tt.factor)
			if r.Width != tt.wantWidth {
				t.Errorf("Width = %.1f, 期望 %.1f", r.Width, tt.wantWidth)
			}
			if r.Height != tt.wantHeight {
				t.Errorf("Height = %.1f, 期望 %.1f", r.Height, tt.wantHeight)
			}
		})
	}
}

// TestRectangle_String 测试 Rectangle.String 实现 fmt.Stringer
func TestRectangle_String(t *testing.T) {
	tests := []struct {
		name           string
		width, height  float64
		wantContains   string
	}{
		{name: "标准矩形", width: 10, height: 5, wantContains: "50.0"},
		{name: "正方形", width: 4, height: 4, wantContains: "16.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rectangle{Width: tt.width, Height: tt.height}
			s := r.String()
			if !strings.Contains(s, tt.wantContains) {
				t.Errorf("String() = %q, 应包含 %q", s, tt.wantContains)
			}
			if !strings.Contains(s, "Rectangle") {
				t.Error("String() 应包含 'Rectangle'")
			}
		})
	}
}

// ============================================================
// 5. 结构体校验（Validate / Greet / SetAge）
// ============================================================

// TestUserInfo_Validate 测试 UserInfo.Validate 校验逻辑
func TestUserInfo_Validate(t *testing.T) {
	tests := []struct {
		name     string
		user     UserInfo
		wantErr  bool
		wantMsg  string // 期望错误消息包含的子串
	}{
		{name: "合法用户", user: UserInfo{Username: "alice", Email: "alice@example.com", Age: 28}, wantErr: false},
		{name: "用户名太短", user: UserInfo{Username: "ab", Email: "test@test.com", Age: 20}, wantErr: true, wantMsg: "3-20"},
		{name: "用户名太长", user: UserInfo{Username: "this-username-is-way-too-long-and-should-fail", Email: "test@test.com", Age: 20}, wantErr: true, wantMsg: "3-20"},
		{name: "无效邮箱（无@）", user: UserInfo{Username: "validuser", Email: "invalid", Age: 20}, wantErr: true, wantMsg: "邮箱"},
		{name: "年龄为负", user: UserInfo{Username: "validuser", Email: "a@b.com", Age: -1}, wantErr: true, wantMsg: "年龄"},
		{name: "年龄过大", user: UserInfo{Username: "validuser", Email: "a@b.com", Age: 151}, wantErr: true, wantMsg: "年龄"},
		{name: "年龄边界 0", user: UserInfo{Username: "validuser", Email: "a@b.com", Age: 0}, wantErr: false},
		{name: "年龄边界 150", user: UserInfo{Username: "validuser", Email: "a@b.com", Age: 150}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("期望错误，但没有得到")
				} else if tt.wantMsg != "" && !strings.Contains(err.Error(), tt.wantMsg) {
					t.Errorf("错误消息 %q 应包含 %q", err.Error(), tt.wantMsg)
				}
			} else {
				if err != nil {
					t.Errorf("期望无错误，得到 %v", err)
				}
			}
		})
	}
}

// TestUserInfo_Greet 测试 UserInfo.Greet 问候语
func TestUserInfo_Greet(t *testing.T) {
	tests := []struct {
		name     string
		user     UserInfo
		wantName string
		wantAge  int
	}{
		{name: "正常用户", user: UserInfo{Username: "Alice", Age: 28}, wantName: "Alice", wantAge: 28},
		{name: "0岁用户", user: UserInfo{Username: "Baby", Age: 0}, wantName: "Baby", wantAge: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			greet := tt.user.Greet()
			if !strings.Contains(greet, tt.wantName) {
				t.Errorf("Greet=%q 应包含用户名 %q", greet, tt.wantName)
			}
		})
	}
}

// TestUserInfo_SetAge 测试指针接收者修改年龄
func TestUserInfo_SetAge(t *testing.T) {
	t.Run("设置合法年龄", func(t *testing.T) {
		u := UserInfo{Username: "test", Email: "a@b.com", Age: 20}
		err := u.SetAge(25)
		if err != nil {
			t.Fatalf("期望无错误，得到 %v", err)
		}
		if u.Age != 25 {
			t.Errorf("Age = %d, 期望 25", u.Age)
		}
	})

	t.Run("设置非法年龄（负值）", func(t *testing.T) {
		u := UserInfo{Username: "test", Email: "a@b.com", Age: 20}
		err := u.SetAge(-5)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
		if u.Age != 20 {
			t.Errorf("非法 SetAge 不应修改原值, Age = %d, 期望 20", u.Age)
		}
	})

	t.Run("设置非法年龄（过大）", func(t *testing.T) {
		u := UserInfo{Username: "test", Email: "a@b.com", Age: 20}
		err := u.SetAge(200)
		if err == nil {
			t.Fatal("期望错误，但没有得到")
		}
	})
}

// ============================================================
// 6. 辅助函数测试
// ============================================================

// TestStringsContains 测试内部的 stringsContains 函数
func TestStringsContains(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		substr  string
		want    bool
	}{
		{name: "包含子串", s: "alice@example.com", substr: "@", want: true},
		{name: "不包含子串", s: "invalid", substr: "@", want: false},
		{name: "空字符串不包含", s: "", substr: "@", want: false},
		{name: "查找空子串", s: "hello", substr: "", want: true},
		{name: "子串在开头", s: "hello world", substr: "hello", want: true},
		{name: "子串在结尾", s: "hello world", substr: "world", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringsContains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("stringsContains(%q, %q) = %t, 期望 %t", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// ============================================================
// 7. 综合场景
// ============================================================

// TestUserInfo_完整生命周期 模拟用户完整使用流程
func TestUserInfo_完整生命周期(t *testing.T) {
	u := UserInfo{Username: "gopher", Email: "gopher@go.dev", Age: 10}

	// Validate
	if err := u.Validate(); err != nil {
		t.Fatalf("初始验证失败: %v", err)
	}

	// Greet
	greet := u.Greet()
	if !strings.Contains(greet, "gopher") {
		t.Errorf("Greet 应包含用户名: %q", greet)
	}

	// SetAge
	if err := u.SetAge(11); err != nil {
		t.Fatalf("SetAge 失败: %v", err)
	}
	if u.Age != 11 {
		t.Errorf("Age = %d, 期望 11", u.Age)
	}

	// 修改后仍合法
	if err := u.Validate(); err != nil {
		t.Errorf("修改后验证失败: %v", err)
	}
}

// TestConfig_综合场景 验证 Config 的序列化/反序列化完整流程
func TestConfig_综合场景(t *testing.T) {
	// 序列化
	cfg := Config{
		Host:    "example.com",
		Port:    443,
		Debug:   true,
		Secret:  "s3cret!",
		Timeout: 60,
		Tags:    []string{"prod", "us-east"},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	// 反序列化回 Config（Secret 在 JSON 中不存在，应为零值）
	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}
	if decoded.Host != "example.com" {
		t.Errorf("Host = %q, 期望 example.com", decoded.Host)
	}
	if decoded.Port != 443 {
		t.Errorf("Port = %d, 期望 443", decoded.Port)
	}
	if decoded.Debug != true {
		t.Error("Debug 应为 true")
	}
	if decoded.Secret != "" {
		t.Errorf("Secret 应为空（json:\"-\" 忽略序列化）, 得到 %q", decoded.Secret)
	}
}