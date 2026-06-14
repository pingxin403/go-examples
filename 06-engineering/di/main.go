// Go 依赖注入（Dependency Injection）示例
//
// 本文件演示 Go 中依赖注入的常见模式（不使用任何第三方 DI 框架）：
//   - 构造函数注入（Constructor Injection）
//   - 接口依赖（Depend on interfaces, not concretions）
//   - 手动组装依赖（Wire up in main）
//   - 隐式 DI vs 服务定位器反模式
//
// 核心思想：一个组件不应该自己创建它的依赖，而应该由外部传入。
// 这样做的好处是：组件更容易测试、更灵活、耦合度更低。
package main

import (
	"errors"
	"fmt"
	"sync"
)

// ============================================================
// 1. 定义依赖接口
// ============================================================

// User 用户实体
type User struct {
	ID    int
	Name  string
	Email string
}

// UserRepository 用户仓库接口 — 依赖倒置原则：抽象不依赖细节
type UserRepository interface {
	// FindByID 根据 ID 查找用户
	FindByID(id int) (*User, error)
	// Save 保存用户
	Save(user *User) error
	// Delete 删除用户
	Delete(id int) error
}

// ============================================================
// 2. 具体实现 A: InMemoryUserRepository（内存实现，适合测试）
// ============================================================

// InMemoryUserRepository 内存版用户仓库实现
// 使用 sync.RWMutex 保证并发安全
type InMemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

// NewInMemoryUserRepository 创建内存用户仓库
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) FindByID(id int) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("用户 %d 不存在", id)
	}
	// 返回副本防止外部修改
	u := *user
	return &u, nil
}

func (r *InMemoryUserRepository) Save(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == 0 {
		// 新用户，分配 ID
		user.ID = r.nextID
		r.nextID++
	}
	// 保存副本
	u := *user
	r.users[user.ID] = &u
	return nil
}

func (r *InMemoryUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("用户 %d 不存在，无法删除", id)
	}
	delete(r.users, id)
	return nil
}

// ============================================================
// 3. 具体实现 B: MySQLUserRepository（模拟数据库实现）
// ============================================================

// MySQLUserRepository 模拟 MySQL 用户仓库
// 真实场景会连接数据库，这里用模拟数据演示
type MySQLUserRepository struct {
	connString string
	data       map[int]*User
}

// NewMySQLUserRepository 创建 MySQL 用户仓库
func NewMySQLUserRepository(connString string) *MySQLUserRepository {
	fmt.Printf("  [MySQL] 连接数据库: %s\n", connString)
	return &MySQLUserRepository{
		connString: connString,
		data:       make(map[int]*User),
	}
}

func (r *MySQLUserRepository) FindByID(id int) (*User, error) {
	fmt.Printf("  [MySQL] 查询用户: id=%d\n", id)
	user, ok := r.data[id]
	if !ok {
		return nil, fmt.Errorf("用户 %d 不存在", id)
	}
	u := *user
	return &u, nil
}

func (r *MySQLUserRepository) Save(user *User) error {
	fmt.Printf("  [MySQL] 保存用户: %+v\n", user)
	if user.ID == 0 {
		user.ID = len(r.data) + 1 // 简化版 ID 生成
	}
	u := *user
	r.data[user.ID] = &u
	return nil
}

func (r *MySQLUserRepository) Delete(id int) error {
	fmt.Printf("  [MySQL] 删除用户: id=%d\n", id)
	if _, ok := r.data[id]; !ok {
		return fmt.Errorf("用户 %d 不存在", id)
	}
	delete(r.data, id)
	return nil
}

// ============================================================
// 4. 构造函数注入 — UserService 依赖 UserRepository
// ============================================================

// UserService 用户服务，依赖 UserRepository 接口
//
// 依赖通过构造函数注入（Constructor Injection），这是最推荐的方式：
//   - 依赖在创建时就被确定，不可变（immutable）
//   - 明确可见，不需要看实现细节
//   - 很容易替换为 mock 实现进行测试
type UserService struct {
	// repo 是接口类型，不依赖具体实现
	repo UserRepository
}

// NewUserService 构造函数注入 — 显式声明依赖
//
// 这是 Go 中最常见的 DI 模式：
//   - 参数列表明确列出了所有依赖
//   - 返回 *UserService，调用者无需知道内部实现
//   - 编译期保证所有依赖都已提供
func NewUserService(repo UserRepository) *UserService {
	if repo == nil {
		// 防御性编程：依赖不能为 nil
		panic("UserRepository 不能为 nil")
	}
	return &UserService{repo: repo}
}

// Register 注册新用户
func (s *UserService) Register(name, email string) (*User, error) {
	if name == "" {
		return nil, errors.New("用户名不能为空")
	}
	if email == "" {
		return nil, errors.New("邮箱不能为空")
	}

	user := &User{Name: name, Email: email}
	if err := s.repo.Save(user); err != nil {
		return nil, fmt.Errorf("保存用户失败: %w", err)
	}
	return user, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int) error {
	return s.repo.Delete(id)
}

// ============================================================
// 5. 服务定位器反模式（Service Locator Anti-Pattern）
// ============================================================

// ServiceLocator 服务定位器 — 反模式演示
//
// 问题：
//   - 隐藏依赖：只看方法签名不知道需要什么
//   - 运行时可能失败：依赖不存在时才能发现
//   - 测试困难：需要 mock 整个定位器
//   - 类型不安全：需要类型断言
type ServiceLocator struct {
	services map[string]interface{}
}

func NewServiceLocator() *ServiceLocator {
	return &ServiceLocator{
		services: make(map[string]interface{}),
	}
}

func (l *ServiceLocator) Register(name string, service interface{}) {
	l.services[name] = service
}

func (l *ServiceLocator) Get(name string) interface{} {
	return l.services[name]
}

// UserServiceWithLocator 使用服务定位器的 UserService（反模式）
// 从方法签名完全看不出它依赖数据库！
type UserServiceWithLocator struct {
	locator *ServiceLocator
}

func NewUserServiceWithLocator(locator *ServiceLocator) *UserServiceWithLocator {
	return &UserServiceWithLocator{locator: locator}
}

func (s *UserServiceWithLocator) Register(name, email string) (*User, error) {
	// 运行时才能发现依赖，而且类型断言可能失败
	repo, ok := s.locator.Get("user_repo").(UserRepository)
	if !ok {
		return nil, errors.New("user_repo 未注册或类型不匹配")
	}
	user := &User{Name: name, Email: email}
	return user, repo.Save(user)
}

// ============================================================
// 6. 手动组装依赖（Wire Up）
// ============================================================

// 在 main 函数中手动组装所有依赖，这是 Go 中最常用的方式。
// 随着项目增长，可迁移到 Google Wire 等代码生成工具。

func demoConstructorInjection() {
	fmt.Println("=== 示例1: 构造函数注入（推荐）===")
	fmt.Println()

	// 方式 A：使用内存实现（适合开发/测试）
	fmt.Println("--- 场景A: InMemory 实现 ---")
	inMemoryRepo := NewInMemoryUserRepository()
	userService := NewUserService(inMemoryRepo) // 构造函数注入

	// 使用服务
	alice, _ := userService.Register("Alice", "alice@example.com")
	fmt.Printf("  注册用户: ID=%d, Name=%s, Email=%s\n", alice.ID, alice.Name, alice.Email)

	bob, _ := userService.Register("Bob", "bob@example.com")
	fmt.Printf("  注册用户: ID=%d, Name=%s, Email=%s\n", bob.ID, bob.Name, bob.Email)

	found, _ := userService.GetUser(1)
	fmt.Printf("  查询用户 1: %+v\n", found)

	// 方式 B：切换为 MySQL 实现（只需改一行）
	fmt.Println()
	fmt.Println("--- 场景B: MySQL 实现（一行切换）---")
	mysqlRepo := NewMySQLUserRepository("user:pass@tcp(localhost:3306)/myapp")
	mysqlService := NewUserService(mysqlRepo) // 同样的构造函数，不同的实现

	mysqlService.Register("Charlie", "charlie@example.com")
	mysqlService.Register("Diana", "diana@example.com")
	mysqlService.GetUser(1)
	fmt.Println()
}

func demoInterfaceBasedDI() {
	fmt.Println("=== 示例2: 面向接口编程 ===")
	fmt.Println()

	// 不同实现可以放在同一个接口变量中
	var repo UserRepository

	// 切换到 InMemory
	repo = NewInMemoryUserRepository()
	testService := NewUserService(repo)
	fmt.Println("  使用 InMemory 实现:")
	testService.Register("Test", "test@test.com")

	// 切换到 MySQL（同一 repo 变量，不同实现）
	repo = NewMySQLUserRepository("test:pass@tcp(test:3306)/testdb")
	testService2 := NewUserService(repo)
	fmt.Println("  使用 MySQL 实现:")
	testService2.Register("Test", "test@test.com")

	fmt.Println()
}

func demoAntiPattern() {
	fmt.Println("=== 示例3: 服务定位器反模式（不推荐）===")
	fmt.Println()

	locator := NewServiceLocator()
	locator.Register("user_repo", NewInMemoryUserRepository())

	badService := NewUserServiceWithLocator(locator)
	user, err := badService.Register("Eve", "eve@example.com")
	if err != nil {
		fmt.Printf("  注册失败: %v\n", err)
	} else {
		fmt.Printf("  注册成功（但依赖不透明）: %+v\n", user)
	}

	fmt.Println()
	fmt.Println("  ⚠️ 服务定位器的问题:")
	fmt.Println("   - 方法签名看不出依赖")
	fmt.Println("   - 运行时才能发现缺少依赖")
	fmt.Println("   - 类型不安全（需要 interface{} 断言）")
	fmt.Println("   - 测试时需要 mock 整个定位器")
	fmt.Println()
}

func demoTestsWithDI() {
	fmt.Println("=== 示例4: DI 让测试变得简单 ===")
	fmt.Println()

	// 测试时使用内存实现，不依赖外部数据库
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	// 测试1: 注册用户
	user, err := service.Register("TestUser", "test@test.com")
	if err != nil {
		fmt.Printf("  ❌ 测试1失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 测试1通过: 注册用户 ID=%d\n", user.ID)
	}

	// 测试2: 查询不存在的用户
	_, err = service.GetUser(999)
	if err != nil {
		fmt.Printf("  ✅ 测试2通过: 查询不存在用户返回错误: %v\n", err)
	}

	// 测试3: 空用户名
	_, err = service.Register("", "empty@test.com")
	if err != nil {
		fmt.Printf("  ✅ 测试3通过: 空用户名被拒绝: %v\n", err)
	}

	// 测试4: 删除用户
	err = service.DeleteUser(1)
	if err != nil {
		fmt.Printf("  ❌ 测试4失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 测试4通过: 用户 1 已删除\n")
	}
}

func main() {
	fmt.Println("========================================")
	fmt.Println("Go 依赖注入（DI）示例")
	fmt.Println("========================================")
	fmt.Println()

	demoConstructorInjection()
	demoInterfaceBasedDI()
	demoAntiPattern()
	demoTestsWithDI()

	fmt.Println("========================================")
	fmt.Println("总结")
	fmt.Println("========================================")
	fmt.Println("推荐做法:")
	fmt.Println("  1. 面向接口编程，依赖抽象而非具体实现")
	fmt.Println("  2. 使用构造函数注入，显式声明依赖")
	fmt.Println("  3. 在 main 中手动组装依赖（Wiring）")
	fmt.Println("  4. 对于大型项目，考虑 Google Wire 等代码生成工具")
	fmt.Println()
	fmt.Println("避免做法:")
	fmt.Println("  1. ❌ 服务定位器（Service Locator）")
	fmt.Println("  2. ❌ 全局变量/单例模式隐藏依赖")
	fmt.Println("  3. ❌ 在组件内部 new 依赖对象")
}