// main_test.go — 对 di 包中可测试函数的完整测试套件
//
// 本文件测试:
//   - InMemoryUserRepository — 内存用户仓库（CRUD）
//   - UserService — 用户服务（Register / GetUser / DeleteUser）
//   - NewUserService — 构造函数注入（nil 保护）
//   - ServiceLocator — 服务定位器（Register / Get）
//   - MySQLUserRepository — MySQL 实现（CRUD）
package main

import (
	"testing"
)

// ============================================================
// 1. InMemoryUserRepository — 内存用户仓库
// ============================================================

// TestNewInMemoryUserRepository 测试创建内存仓库的初始状态
func TestNewInMemoryUserRepository(t *testing.T) {
	repo := NewInMemoryUserRepository()
	if repo == nil {
		t.Fatal("NewInMemoryUserRepository() 返回了 nil")
	}

	// 初始时查询不存在的用户应返回错误
	_, err := repo.FindByID(1)
	if err == nil {
		t.Error("空仓库查询用户 1 应返回错误")
	}
}

// TestInMemoryRepo_SaveAndFind 测试保存用户后能正确查询
func TestInMemoryRepo_SaveAndFind(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// 保存新用户（ID=0 时自动分配 ID）
	user := &User{Name: "Alice", Email: "alice@example.com"}
	if err := repo.Save(user); err != nil {
		t.Fatalf("保存用户失败: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("自动分配的 ID = %d; 期望 1", user.ID)
	}

	// 查询刚保存的用户
	found, err := repo.FindByID(1)
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if found.Name != "Alice" {
		t.Errorf("用户名称 = %q; 期望 %q", found.Name, "Alice")
	}
	if found.Email != "alice@example.com" {
		t.Errorf("用户邮箱 = %q; 期望 %q", found.Email, "alice@example.com")
	}
}

// TestInMemoryRepo_SaveMultiple 测试多用户保存与 ID 自增
func TestInMemoryRepo_SaveMultiple(t *testing.T) {
	repo := NewInMemoryUserRepository()

	users := []*User{
		{Name: "Alice", Email: "alice@test.com"},
		{Name: "Bob", Email: "bob@test.com"},
		{Name: "Charlie", Email: "charlie@test.com"},
	}

	for i, u := range users {
		if err := repo.Save(u); err != nil {
			t.Fatalf("保存用户 %d 失败: %v", i+1, err)
		}
		if u.ID != i+1 {
			t.Errorf("用户 %q 的 ID = %d; 期望 %d", u.Name, u.ID, i+1)
		}
	}

	// 验证所有用户能正确查询
	for i, u := range users {
		found, err := repo.FindByID(i + 1)
		if err != nil {
			t.Errorf("查询用户 %d 失败: %v", i+1, err)
			continue
		}
		if found.Name != u.Name {
			t.Errorf("FindByID(%d).Name = %q; 期望 %q", i+1, found.Name, u.Name)
		}
	}
}

// TestInMemoryRepo_FindByID_NotFound 测试查询不存在的用户
func TestInMemoryRepo_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	_, err := repo.FindByID(999)
	if err == nil {
		t.Error("查询不存在的用户应返回错误")
	}
}

// TestInMemoryRepo_UpdateExistingUser 测试更新已有用户（ID>0）
func TestInMemoryRepo_UpdateExistingUser(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// 保存新用户
	user := &User{Name: "Alice", Email: "alice@test.com"}
	repo.Save(user)

	// 更新用户（ID 不变，修改名称）
	user.Name = "Alice Updated"
	if err := repo.Save(user); err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	// 验证更新生效
	found, _ := repo.FindByID(1)
	if found.Name != "Alice Updated" {
		t.Errorf("更新后名称 = %q; 期望 %q", found.Name, "Alice Updated")
	}
}

// TestInMemoryRepo_Delete 测试删除用户
func TestInMemoryRepo_Delete(t *testing.T) {
	repo := NewInMemoryUserRepository()

	// 保存两个用户
	repo.Save(&User{Name: "Alice", Email: "a@test.com"})
	repo.Save(&User{Name: "Bob", Email: "b@test.com"})

	// 删除 Alice (ID=1)
	if err := repo.Delete(1); err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	// 验证 Alice 已被删除
	_, err := repo.FindByID(1)
	if err == nil {
		t.Error("删除后查询用户应返回错误")
	}

	// 验证 Bob 仍在
	bob, err := repo.FindByID(2)
	if err != nil {
		t.Errorf("查询用户 2 失败: %v", err)
	}
	if bob.Name != "Bob" {
		t.Errorf("剩余用户名称 = %q; 期望 %q", bob.Name, "Bob")
	}
}

// TestInMemoryRepo_Delete_NotFound 测试删除不存在的用户
func TestInMemoryRepo_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryUserRepository()

	err := repo.Delete(999)
	if err == nil {
		t.Error("删除不存在的用户应返回错误")
	}
}

// ============================================================
// 2. UserService — 构造函数注入的用户服务
// ============================================================

// TestNewUserService 测试创建 UserService
func TestNewUserService(t *testing.T) {
	t.Run("传入有效 repo", func(t *testing.T) {
		repo := NewInMemoryUserRepository()
		svc := NewUserService(repo)
		if svc == nil {
			t.Fatal("NewUserService 返回了 nil")
		}
	})

	t.Run("传入 nil 应 panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("传入 nil repo 应触发 panic")
			}
		}()
		NewUserService(nil)
	})
}

// TestUserService_Register 测试用户注册
func TestUserService_Register(t *testing.T) {
	repo := NewInMemoryUserRepository()
	svc := NewUserService(repo)

	t.Run("正常注册", func(t *testing.T) {
		user, err := svc.Register("Alice", "alice@example.com")
		if err != nil {
			t.Fatalf("注册用户失败: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("注册用户名称 = %q; 期望 %q", user.Name, "Alice")
		}
		if user.Email != "alice@example.com" {
			t.Errorf("注册用户邮箱 = %q; 期望 %q", user.Email, "alice@example.com")
		}
		if user.ID == 0 {
			t.Error("注册用户应获得非零 ID")
		}
	})

	t.Run("空用户名被拒绝", func(t *testing.T) {
		_, err := svc.Register("", "empty@test.com")
		if err == nil {
			t.Error("空用户名应返回错误")
		}
	})

	t.Run("空邮箱被拒绝", func(t *testing.T) {
		_, err := svc.Register("NoEmail", "")
		if err == nil {
			t.Error("空邮箱应返回错误")
		}
	})

	t.Run("注册多个用户 ID 自增", func(t *testing.T) {
		user1, _ := svc.Register("Alice", "a@test.com")
		user2, _ := svc.Register("Bob", "b@test.com")
		user3, _ := svc.Register("Charlie", "c@test.com")

		if user2.ID != user1.ID+1 {
			t.Errorf("用户 2 的 ID 应 = %d, 得到 %d", user1.ID+1, user2.ID)
		}
		if user3.ID != user2.ID+1 {
			t.Errorf("用户 3 的 ID 应 = %d, 得到 %d", user2.ID+1, user3.ID)
		}
	})
}

// TestUserService_GetUser 测试获取用户
func TestUserService_GetUser(t *testing.T) {
	repo := NewInMemoryUserRepository()
	svc := NewUserService(repo)

	t.Run("获取已存在的用户", func(t *testing.T) {
		registered, _ := svc.Register("Alice", "a@test.com")
		found, err := svc.GetUser(registered.ID)
		if err != nil {
			t.Fatalf("GetUser 失败: %v", err)
		}
		if found.Name != "Alice" {
			t.Errorf("查询用户名称 = %q; 期望 %q", found.Name, "Alice")
		}
	})

	t.Run("获取不存在的用户", func(t *testing.T) {
		_, err := svc.GetUser(999)
		if err == nil {
			t.Error("获取不存在的用户应返回错误")
		}
	})
}

// TestUserService_DeleteUser 测试删除用户
func TestUserService_DeleteUser(t *testing.T) {
	repo := NewInMemoryUserRepository()
	svc := NewUserService(repo)

	t.Run("删除已存在的用户", func(t *testing.T) {
		user, _ := svc.Register("Alice", "a@test.com")
		err := svc.DeleteUser(user.ID)
		if err != nil {
			t.Fatalf("DeleteUser 失败: %v", err)
		}
		// 删除后不可查询
		_, err = svc.GetUser(user.ID)
		if err == nil {
			t.Error("删除后 GetUser 应返回错误")
		}
	})

	t.Run("删除不存在的用户", func(t *testing.T) {
		err := svc.DeleteUser(999)
		if err == nil {
			t.Error("删除不存在的用户应返回错误")
		}
	})
}

// ============================================================
// 3. ServiceLocator — 服务定位器（反模式演示）
// ============================================================

// TestServiceLocator 测试服务定位器的注册与获取
func TestServiceLocator(t *testing.T) {
	locator := NewServiceLocator()

	t.Run("注册并获取服务", func(t *testing.T) {
		repo := NewInMemoryUserRepository()
		locator.Register("user_repo", repo)

		got := locator.Get("user_repo")
		if got == nil {
			t.Fatal("Get 返回了 nil")
		}
		// 验证类型转换
		_, ok := got.(UserRepository)
		if !ok {
			t.Error("获取的服务无法转换为 UserRepository 接口")
		}
	})

	t.Run("获取未注册的服务", func(t *testing.T) {
		got := locator.Get("nonexistent")
		if got != nil {
			t.Errorf("获取未注册服务应返回 nil, 得到 %v", got)
		}
	})
}

// TestUserServiceWithLocator 测试基于服务定位器的 UserService
func TestUserServiceWithLocator(t *testing.T) {
	locator := NewServiceLocator()
	locator.Register("user_repo", NewInMemoryUserRepository())
	svc := NewUserServiceWithLocator(locator)

	t.Run("注册用户", func(t *testing.T) {
		user, err := svc.Register("Eve", "eve@example.com")
		if err != nil {
			t.Fatalf("Register 失败: %v", err)
		}
		if user.Name != "Eve" {
			t.Errorf("用户名称 = %q; 期望 %q", user.Name, "Eve")
		}
	})
}

// ============================================================
// 4. MySQLUserRepository — MySQL 实现
// ============================================================

// TestMySQLUserRepository 测试 MySQLUserRepository 的基本 CRUD
func TestMySQLUserRepository(t *testing.T) {
	repo := NewMySQLUserRepository("user:pass@tcp(localhost:3306)/testdb")

	t.Run("保存并查询用户", func(t *testing.T) {
		user := &User{Name: "Charlie", Email: "charlie@test.com"}
		if err := repo.Save(user); err != nil {
			t.Fatalf("Save 失败: %v", err)
		}
		if user.ID == 0 {
			t.Error("保存后用户应获得 ID")
		}

		found, err := repo.FindByID(user.ID)
		if err != nil {
			t.Fatalf("FindByID 失败: %v", err)
		}
		if found.Name != "Charlie" {
			t.Errorf("查询用户名称 = %q; 期望 %q", found.Name, "Charlie")
		}
	})

	t.Run("查询不存在的用户", func(t *testing.T) {
		_, err := repo.FindByID(999)
		if err == nil {
			t.Error("查询不存在的用户应返回错误")
		}
	})

	t.Run("删除用户", func(t *testing.T) {
		user := &User{Name: "Diana", Email: "diana@test.com"}
		repo.Save(user)
		if err := repo.Delete(user.ID); err != nil {
			t.Fatalf("Delete 失败: %v", err)
		}
		// 删除后不可查询
		_, err := repo.FindByID(user.ID)
		if err == nil {
			t.Error("删除后 FindByID 应返回错误")
		}
	})

	t.Run("删除不存在的用户", func(t *testing.T) {
		err := repo.Delete(999)
		if err == nil {
			t.Error("删除不存在的用户应返回错误")
		}
	})
}