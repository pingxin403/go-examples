package main

import (
	"fmt"
	"reflect"
	"strings"
)

/*
 * ===========================
 * Go 反射（reflect 包）示例
 * ===========================
 *
 * 反射是 Go 语言的强大机制，允许程序在运行时检查、修改自身的结构和行为。
 * 适用场景：JSON 序列化/反序列化、ORM、依赖注入、通用校验等。
 *
 * 核心函数：
 *   reflect.TypeOf(v)  → 获取类型信息
 *   reflect.ValueOf(v) → 获取值信息
 */

// ============ 1. 用于演示的结构体 ============

// User 用户结构体，演示反射读取字段和标签
type User struct {
	ID    int    `json:"id"    validate:"required"`
	Name  string `json:"name"  validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age"   validate:"min=0,max=150"`
	Phone string `json:"phone,omitempty"`
}

// ============ 2. 用于演示的动态函数调用 ============

// Greet 一个普通函数，稍后通过反射动态调用
func Greet(name string, age int) string {
	return fmt.Sprintf("你好，%s！你今年 %d 岁。", name, age)
}

// ============ 3. 实用：基于 struct tag 的通用校验器 ============

// ValidationError 校验错误
type ValidationError struct {
	Field   string
	Tag     string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] 标签 %q 校验失败: %s", e.Field, e.Tag, e.Message)
}

// ValidateStruct 基于 struct tag 的通用校验器
// 解析 validate 标签，对字段执行基本的校验规则
func ValidateStruct(s interface{}) []error {
	var errs []error

	val := reflect.ValueOf(s)
	// 必须传入结构体指针才能获取到导出的字段
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return append(errs, fmt.Errorf("expected struct, got %s", val.Kind()))
	}

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// 获取 validate 标签
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// 获取 json 标签用于错误信息
		jsonTag := field.Tag.Get("json")
		fieldName := jsonTag
		if idx := strings.Index(jsonTag, ","); idx > 0 {
			fieldName = jsonTag[:idx]
		}
		if fieldName == "" {
			fieldName = field.Name
		}

		// 解析 validate 标签规则（逗号分隔）
		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			switch {
			case rule == "required":
				// 检查零值：对于指针检查 nil，其他类型检查零值
				if fieldVal.Kind() == reflect.Ptr {
					if fieldVal.IsNil() {
						errs = append(errs, ValidationError{
							Field: fieldName, Tag: "required",
							Message: "字段不能为空",
						})
					}
				} else if fieldVal.IsZero() {
					errs = append(errs, ValidationError{
						Field: fieldName, Tag: "required",
						Message: "字段不能为默认零值",
					})
				}

			case strings.HasPrefix(rule, "min="):
				var min int
				fmt.Sscanf(rule, "min=%d", &min)
				switch fieldVal.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if fieldVal.Int() < int64(min) {
						errs = append(errs, ValidationError{
							Field: fieldName, Tag: "min",
							Message: fmt.Sprintf("值 %d 小于最小值 %d", fieldVal.Int(), min),
						})
					}
				case reflect.String:
					if len(fieldVal.String()) < min {
						errs = append(errs, ValidationError{
							Field: fieldName, Tag: "min",
							Message: fmt.Sprintf("长度 %d 小于最小长度 %d", len(fieldVal.String()), min),
						})
					}
				}

			case strings.HasPrefix(rule, "max="):
				var max int
				fmt.Sscanf(rule, "max=%d", &max)
				switch fieldVal.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if fieldVal.Int() > int64(max) {
						errs = append(errs, ValidationError{
							Field: fieldName, Tag: "max",
							Message: fmt.Sprintf("值 %d 大于最大值 %d", fieldVal.Int(), max),
						})
					}
				case reflect.String:
					if len(fieldVal.String()) > max {
						errs = append(errs, ValidationError{
							Field: fieldName, Tag: "max",
							Message: fmt.Sprintf("长度 %d 大于最大长度 %d", len(fieldVal.String()), max),
						})
					}
				}

			case rule == "email":
				if fieldVal.Kind() == reflect.String && fieldVal.String() != "" {
					s := fieldVal.String()
					if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
						errs = append(errs, ValidationError{
							Field: fieldName, Tag: "email",
							Message: fmt.Sprintf("%q 不是有效的 email 地址", s),
						})
					}
				}
			}
		}
	}

	return errs
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Go 反射（reflect 包）完整示例")
	fmt.Println("========================================")

	// ====================================================
	// 1. 检查变量的类型和种类
	// ====================================================
	fmt.Println("\n--- 1. 检查类型 (Type) 和种类 (Kind) ---")

	var x float64 = 3.14159
	t := reflect.TypeOf(x)
	v := reflect.ValueOf(x)

	fmt.Printf("变量 x = %v\n", x)
	fmt.Printf("Type:  %v\n", t)               // float64
	fmt.Printf("Kind:  %v\n", t.Kind())        // float64（和 Type 相同，因为 float64 是底层类型）
	fmt.Printf("Value: %v\n", v)               // 3.14159
	fmt.Printf("Value.Type(): %v\n", v.Type()) // float64

	// 注意：Type 和 Kind 的区别
	// Type 是具体的类型（如 User、*os.File）
	// Kind 是底层种类（如 struct、ptr、slice）
	var u User
	ut := reflect.TypeOf(u)
	fmt.Printf("\nUser 的 Type:  %v\n", ut)   // main.User
	fmt.Printf("User 的 Kind:  %v\n", ut.Kind()) // struct

	var up *User
	upt := reflect.TypeOf(up)
	fmt.Printf("*User 的 Type: %v\n", upt)   // *main.User
	fmt.Printf("*User 的 Kind: %v\n", upt.Kind()) // ptr（不是 main.User！）

	// ====================================================
	// 2. 读取结构体字段和标签
	// ====================================================
	fmt.Println("\n--- 2. 读取结构体字段和标签 ---")

	user := User{
		ID:    1001,
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   28,
	}

	userType := reflect.TypeOf(user)
	fmt.Printf("结构体 %s 共有 %d 个字段:\n\n", userType.Name(), userType.NumField())

	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		fmt.Printf("  字段 %d: %s\n", i, field.Name)
		fmt.Printf("    类型: %v\n", field.Type)
		fmt.Printf("    JSON 标签: %q\n", field.Tag.Get("json"))
		fmt.Printf("    validate 标签: %q\n", field.Tag.Get("validate"))
		fmt.Printf("    是否是导出字段: %v\n", field.IsExported())
		fmt.Println()
	}

	// ====================================================
	// 3. 通过反射设置字段值
	// ====================================================
	fmt.Println("--- 3. 通过反射设置字段值 ---")

	// 注意：要设置值，必须传入指针
	userPtr := &user
	rv := reflect.ValueOf(userPtr).Elem()

	nameField := rv.FieldByName("Name")
	if nameField.IsValid() && nameField.CanSet() {
		nameField.SetString("李四")
	}

	ageField := rv.FieldByName("Age")
	if ageField.IsValid() && ageField.CanSet() {
		ageField.SetInt(30)
	}

	fmt.Printf("修改后: user.Name = %q, user.Age = %d\n\n", user.Name, user.Age)

	// ====================================================
	// 4. 动态调用函数
	// ====================================================
	fmt.Println("--- 4. 动态调用函数 ---")

	funcValue := reflect.ValueOf(Greet)
	params := []reflect.Value{
		reflect.ValueOf("王五"),
		reflect.ValueOf(25),
	}
	results := funcValue.Call(params)
	fmt.Println("动态调用 Greet 的结果:", results[0].Interface())

	// ====================================================
	// 5. 动态创建实例
	// ====================================================
	fmt.Println("\n--- 5. 动态创建实例 ---")

	// 通过类型动态创建
	newUserType := reflect.TypeOf(User{})
	newUserPtr := reflect.New(newUserType) // 返回 *User
	newUser := newUserPtr.Elem()           // 获取 User 值

	newUser.FieldByName("ID").SetInt(2002)
	newUser.FieldByName("Name").SetString("赵六")
	newUser.FieldByName("Email").SetString("zhaoliu@example.com")
	newUser.FieldByName("Age").SetInt(35)

	fmt.Printf("动态创建的 User: %+v\n\n", newUser.Interface())

	// 动态创建 slice
	sliceType := reflect.SliceOf(newUserType)
	slice := reflect.MakeSlice(sliceType, 0, 3)
	slice = reflect.Append(slice, newUser)
	slice = reflect.Append(slice, rv) // 添加原始的 user

	fmt.Printf("动态创建的 User 切片，长度: %d\n", slice.Len())
	for i := 0; i < slice.Len(); i++ {
		fmt.Printf("  [%d] %+v\n", i, slice.Index(i).Interface())
	}

	// ====================================================
	// 6. 实用示例：通用 JSON 校验器
	// ====================================================
	fmt.Println("\n--- 6. 实用示例：通用 Validator（基于 struct tag）---")

	// 测试合法的 user
	validUser := User{
		ID:    1001,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}
	errs := ValidateStruct(validUser)
	if len(errs) == 0 {
		fmt.Println("✅ validUser 校验通过")
	} else {
		for _, e := range errs {
			fmt.Println("  ", e)
		}
	}

	// 测试有问题的 user
	invalidUser := User{
		ID:    0,     // required, 零值
		Name:  "A",  // min=2 不满足
		Email: "bademail", // email 格式错误
		Age:   200,  // max=150 不满足
	}
	errs = ValidateStruct(invalidUser)
	fmt.Println("\n❌ invalidUser 校验结果:")
	for _, e := range errs {
		fmt.Println("  ", e)
	}

	fmt.Println("\n========================================")
	fmt.Println("  反射总结")
	fmt.Println("========================================")
	fmt.Println("✅ 优点：灵活，适用于通用框架和库开发")
	fmt.Println("⚠️  缺点：")
	fmt.Println("  1. 性能开销大（比直接调用慢 10-100 倍）")
	fmt.Println("  2. 编译期无法发现类型错误（运行时 panic）")
	fmt.Println("  3. 代码可读性和可维护性降低")
	fmt.Println("  4. 某些反射操作需要可达地址（CanAddr）或可设置（CanSet）")
	fmt.Println("💡 建议：普通业务代码尽量少用反射，框架/库代码中合理使用")
}