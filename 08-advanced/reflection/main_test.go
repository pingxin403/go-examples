package main

import (
	"testing"
)

// ============ Greet 函数测试 ============

// TestGreet 测试 Greet 函数的正常调用
func TestGreet(t *testing.T) {
	got := Greet("张三", 25)
	want := "你好，张三！你今年 25 岁。"
	if got != want {
		t.Errorf("Greet() = %q, 期望 %q", got, want)
	}
}

// TestGreet_EmptyName 测试空名字场景
func TestGreet_EmptyName(t *testing.T) {
	got := Greet("", 30)
	want := "你好，！你今年 30 岁。"
	if got != want {
		t.Errorf("Greet() = %q, 期望 %q", got, want)
	}
}

// ============ ValidateStruct 函数测试 ============

// TestValidateStruct_ValidUser 测试合法用户通过校验
func TestValidateStruct_ValidUser(t *testing.T) {
	user := User{
		ID:    1001,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}
	errs := ValidateStruct(user)
	if len(errs) != 0 {
		t.Errorf("合法用户应无校验错误，得到 %d 个错误: %v", len(errs), errs)
	}
}

// TestValidateStruct_ZeroValueID 测试 ID 为 0（required 失败）
func TestValidateStruct_ZeroValueID(t *testing.T) {
	user := User{
		ID:    0,
		Name:  "Bob",
		Email: "bob@example.com",
		Age:   30,
	}
	errs := ValidateStruct(user)
	if len(errs) == 0 {
		t.Fatal("期望校验错误，但没有得到")
	}
	found := false
	for _, e := range errs {
		if e.Error() == `[id] 标签 "required" 校验失败: 字段不能为默认零值` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望找到 id 字段的 required 错误，实际错误: %v", errs)
	}
}

// TestValidateStruct_NameTooShort 测试名字长度不足（min=2）
func TestValidateStruct_NameTooShort(t *testing.T) {
	user := User{
		ID:    1,
		Name:  "A",
		Email: "a@b.com",
		Age:   20,
	}
	errs := ValidateStruct(user)
	found := false
	for _, e := range errs {
		if e.Error() == `[name] 标签 "min" 校验失败: 长度 1 小于最小长度 2` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望找到 name 字段的 min 错误，实际错误: %v", errs)
	}
}

// TestValidateStruct_InvalidEmail 测试无效 email
func TestValidateStruct_InvalidEmail(t *testing.T) {
	user := User{
		ID:    1,
		Name:  "Test",
		Email: "bademail",
		Age:   25,
	}
	errs := ValidateStruct(user)
	found := false
	for _, e := range errs {
		if e.Error() == `[email] 标签 "email" 校验失败: "bademail" 不是有效的 email 地址` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望找到 email 字段的 email 错误，实际错误: %v", errs)
	}
}

// TestValidateStruct_AgeOverMax 测试年龄超上限
func TestValidateStruct_AgeOverMax(t *testing.T) {
	user := User{
		ID:    1,
		Name:  "Test",
		Email: "test@test.com",
		Age:   200,
	}
	errs := ValidateStruct(user)
	found := false
	for _, e := range errs {
		if e.Error() == `[age] 标签 "max" 校验失败: 值 200 大于最大值 150` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望找到 age 字段的 max 错误，实际错误: %v", errs)
	}
}

// TestValidateStruct_AllErrors 测试同时触发所有校验错误
func TestValidateStruct_AllErrors(t *testing.T) {
	invalidUser := User{
		ID:    0,
		Name:  "A",
		Email: "bademail",
		Age:   200,
	}
	errs := ValidateStruct(invalidUser)
	// 期望 4 个错误：id required, name min, email email, age max
	if len(errs) < 4 {
		t.Errorf("期望至少 4 个错误，得到 %d 个: %v", len(errs), errs)
	}
}

// TestValidateStruct_NotStruct 测试非结构体输入
func TestValidateStruct_NotStruct(t *testing.T) {
	errs := ValidateStruct("not a struct")
	if len(errs) == 0 {
		t.Fatal("期望非结构体输入报错，但没有得到")
	}
}

// TestValidateStruct_PtrToStruct 测试结构体指针输入
func TestValidateStruct_PtrToStruct(t *testing.T) {
	user := &User{
		ID:    1001,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}
	errs := ValidateStruct(user)
	if len(errs) != 0 {
		t.Errorf("指针传入的合法用户应无错误，得到: %v", errs)
	}
}

// TestValidateStruct_ZeroEmail 测试空 email（无 validate 标签检查时不应报 email 格式错）
func TestValidateStruct_ZeroEmail(t *testing.T) {
	user := User{
		ID:   1,
		Name: "Bob",
		Age:  25,
		// Email 留空
	}
	errs := ValidateStruct(user)
	// email 字段为空字符串，但 required 标签不在 email 上，应无 email 格式错误
	for _, e := range errs {
		if e.Error() == `[email] 标签 "email" 校验失败: "" 不是有效的 email 地址` {
			t.Errorf("空 email 不应报 email 格式错误: %v", e)
		}
	}
}

// TestValidateStruct_EmailWithoutDot 测试 email 不含点
func TestValidateStruct_EmailWithoutDot(t *testing.T) {
	user := User{
		ID:    1,
		Name:  "Test",
		Email: "test@test",
		Age:   25,
	}
	errs := ValidateStruct(user)
	found := false
	for _, e := range errs {
		if e.Error() == `[email] 标签 "email" 校验失败: "test@test" 不是有效的 email 地址` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望 email（缺少点号）报错，实际错误: %v", errs)
	}
}

// TestValidateStruct_NegativeAge 测试负值年龄（无 min 约束时不应报错）
func TestValidateStruct_NegativeAge(t *testing.T) {
	user := User{
		ID:    1,
		Name:  "Test",
		Email: "test@test.com",
		Age:   -1,
	}
	errs := ValidateStruct(user)
	for _, e := range errs {
		t.Logf("错误: %v", e)
	}
	// age 的 validate 标签是 min=0,max=150，-1 不应触发 max 或 min 之外的规则
	// -1 小于 min=0，所以应该有一个 min 错误
	foundMin := false
	for _, e := range errs {
		if e.Error() == `[age] 标签 "min" 校验失败: 值 -1 小于最小值 0` {
			foundMin = true
			break
		}
	}
	if !foundMin {
		t.Errorf("期望 age < 0 触发 min 错误，实际错误: %v", errs)
	}
}