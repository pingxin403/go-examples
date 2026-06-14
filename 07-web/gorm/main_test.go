// main_test.go — GORM ORM 单元测试
//
// 测试内容：
//   - SQLite 内存数据库连接
//   - AutoMigrate 自动迁移
//   - 基础 CRUD（Create, Read, Update, Delete）
//   - 关联关系（Has Many / Belongs To）
//   - 预加载（Preload）
//   - 事务
//   - 原生 SQL
//   - GORM 表达式

package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB 创建测试用的 SQLite 内存数据库，自动迁移表结构
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试中关闭日志输出
	})
	if err != nil {
		t.Fatalf("连接 SQLite 内存数据库失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&User{}, &Order{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}

	return db
}

// createTestUsers 创建测试用户数据
func createTestUsers(t *testing.T, db *gorm.DB) {
	t.Helper()
	users := []User{
		{Name: "张三", Email: "zhangsan@example.com", Age: 28},
		{Name: "李四", Email: "lisi@example.com", Age: 35},
		{Name: "王五", Email: "wangwu@example.com", Age: 22},
	}
	for _, u := range users {
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("创建测试用户失败: %v", err)
		}
	}
}

// ---------- 测试用例 ----------

// TestDBConnection 测试数据库连接
func TestDBConnection(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("连接 SQLite 内存数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层 *sql.DB 失败: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("数据库 Ping 失败: %v", err)
	}

	sqlDB.Close()
}

// TestAutoMigrate 测试自动迁移
func TestAutoMigrate(t *testing.T) {
	db := setupTestDB(t)

	// 验证表已创建 — 尝试查询应返回空切片而非错误
	var users []User
	if err := db.Find(&users).Error; err != nil {
		t.Fatalf("查询用户表失败: %v", err)
	}

	var orders []Order
	if err := db.Find(&orders).Error; err != nil {
		t.Fatalf("查询订单表失败: %v", err)
	}

	// 验证 Order 表名自定义
	if (Order{}).TableName() != "orders" {
		t.Errorf("预期 Order 表名为 orders，实际得到 %s", (Order{}).TableName())
	}
}

// TestCreateUser 测试创建用户
func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)

	t.Run("创建单个用户", func(t *testing.T) {
		user := User{Name: "测试用户", Email: "test@example.com", Age: 25}
		result := db.Create(&user)

		if result.Error != nil {
			t.Fatalf("创建用户失败: %v", result.Error)
		}
		if user.ID == 0 {
			t.Error("创建后用户 ID 不应为 0")
		}
		if result.RowsAffected != 1 {
			t.Errorf("预期影响 1 行，实际影响 %d", result.RowsAffected)
		}
	})

	t.Run("创建用户 — 缺失必填字段", func(t *testing.T) {
		user := User{Email: "noname@example.com", Age: 20}
		result := db.Create(&user)
		if result.Error != nil {
			t.Logf("不含姓名的用户创建结果（Name 字段非空约束）: %v", result.Error)
		}
	})

	t.Run("批量创建用户", func(t *testing.T) {
		batchUsers := []User{
			{Name: "赵六", Email: "zhaoliu@example.com", Age: 30},
			{Name: "钱七", Email: "qianqi@example.com", Age: 25},
		}
		result := db.Create(&batchUsers)
		if result.Error != nil {
			t.Fatalf("批量创建用户失败: %v", result.Error)
		}
		if result.RowsAffected != 2 {
			t.Errorf("预期影响 2 行，实际影响 %d", result.RowsAffected)
		}
		if batchUsers[0].ID == 0 || batchUsers[1].ID == 0 {
			t.Error("批量创建后用户 ID 不应为 0")
		}
	})
}

// TestReadUser 测试用户查询
func TestReadUser(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	t.Run("按主键查询", func(t *testing.T) {
		var user User
		if err := db.First(&user, 1).Error; err != nil {
			t.Fatalf("First 查询失败: %v", err)
		}
		if user.Name != "张三" {
			t.Errorf("预期 Name=张三，实际得到 %s", user.Name)
		}
	})

	t.Run("按条件查询 — Where", func(t *testing.T) {
		var users []User
		db.Where("age > ?", 25).Find(&users)
		if len(users) != 2 {
			t.Errorf("预期 2 个年龄 > 25 的用户，实际得到 %d", len(users))
		}
	})

	t.Run("按 Struct 条件查询", func(t *testing.T) {
		var user User
		if err := db.Where(&User{Name: "李四"}).First(&user).Error; err != nil {
			t.Fatalf("Struct 条件查询失败: %v", err)
		}
		if user.Email != "lisi@example.com" {
			t.Errorf("预期 Email=lisi@example.com，实际得到 %s", user.Email)
		}
	})

	t.Run("按 Map 条件查询", func(t *testing.T) {
		var users []User
		db.Where(map[string]interface{}{"age": 22}).Find(&users)
		if len(users) != 1 {
			t.Errorf("预期 1 个 age=22 的用户，实际得到 %d", len(users))
		}
	})

	t.Run("查询不存在的记录", func(t *testing.T) {
		var user User
		result := db.First(&user, 999)
		if result.Error != gorm.ErrRecordNotFound {
			t.Errorf("预期 ErrRecordNotFound，实际得到 %v", result.Error)
		}
	})
}

// TestUpdateUser 测试用户更新
func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	t.Run("更新单个字段", func(t *testing.T) {
		result := db.Model(&User{}).Where("name = ?", "王五").Update("age", 23)
		if result.Error != nil {
			t.Fatalf("更新失败: %v", result.Error)
		}
		if result.RowsAffected != 1 {
			t.Errorf("预期影响 1 行，实际影响 %d", result.RowsAffected)
		}

		var user User
		db.Where("name = ?", "王五").First(&user)
		if user.Age != 23 {
			t.Errorf("预期 age=23，实际得到 %d", user.Age)
		}
	})

	t.Run("更新多个字段", func(t *testing.T) {
		result := db.Model(&User{}).Where("name = ?", "张三").Updates(User{Name: "张三丰", Age: 29})
		if result.Error != nil {
			t.Fatalf("更新失败: %v", result.Error)
		}

		var user User
		db.Where("name = ?", "张三丰").First(&user)
		if user.Age != 29 {
			t.Errorf("预期 age=29，实际得到 %d", user.Age)
		}
	})

	t.Run("更新不存在的记录", func(t *testing.T) {
		result := db.Where("name = ?", "不存在").Delete(&User{})
		if result.Error != nil {
			t.Fatalf("删除失败: %v", result.Error)
		}
		// RowsAffected 应为 0
	})
}

// TestDeleteUser 测试用户删除（软删除）
func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	t.Run("软删除用户", func(t *testing.T) {
		result := db.Where("name = ?", "王五").Delete(&User{})
		if result.Error != nil {
			t.Fatalf("软删除失败: %v", result.Error)
		}
		if result.RowsAffected != 1 {
			t.Errorf("预期影响 1 行，实际影响 %d", result.RowsAffected)
		}
	})

	t.Run("软删除后查询不到", func(t *testing.T) {
		var user User
		err := db.Where("name = ?", "王五").First(&user).Error
		if err != gorm.ErrRecordNotFound {
			t.Errorf("软删除后应查不到记录，实际得到 %v", err)
		}
	})

	t.Run("Unscoped 查询包含软删除记录", func(t *testing.T) {
		var allUsers []User
		db.Unscoped().Find(&allUsers)
		if len(allUsers) != 3 {
			t.Errorf("Unscoped 查询预期 3 条记录（含已删除），实际得到 %d", len(allUsers))
		}
	})

	t.Run("硬删除（Unscoped Delete）", func(t *testing.T) {
		db.Unscoped().Where("name = ?", "王五").Delete(&User{})

		// Unscoped 查询也不应再有该记录
		var users []User
		db.Unscoped().Find(&users)
		if len(users) != 2 {
			t.Errorf("硬删除后预期 2 条记录，实际得到 %d", len(users))
		}
	})
}

// TestOrderTableName 测试 Order 自定义表名
func TestOrderTableName(t *testing.T) {
	order := Order{}
	if order.TableName() != "orders" {
		t.Errorf("Order.TableName() 应返回 'orders'，实际返回 '%s'", order.TableName())
	}
}

// TestRelationHasMany 测试 Has Many 关联
func TestRelationHasMany(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	// 获取第一个用户并创建订单
	var user User
	db.First(&user)

	orders := []Order{
		{UserID: user.ID, Product: "Go 语言圣经", Price: 89.00, Quantity: 1, Status: "paid"},
		{UserID: user.ID, Product: "机械键盘", Price: 399.00, Quantity: 2, Status: "pending"},
	}
	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("创建订单失败: %v", err)
		}
	}

	// Preload 查询用户的订单
	var userWithOrders User
	db.Preload("Orders").First(&userWithOrders, user.ID)

	if len(userWithOrders.Orders) != 2 {
		t.Errorf("预期 2 笔订单，实际得到 %d", len(userWithOrders.Orders))
	}
	if userWithOrders.Orders[0].Product == "" {
		t.Error("订单商品名不应为空")
	}
}

// TestRelationBelongsTo 测试 Belongs To 关联
func TestRelationBelongsTo(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	var user User
	db.First(&user)
	db.Create(&Order{UserID: user.ID, Product: "显示器", Price: 159.00, Quantity: 1, Status: "shipped"})

	// Preload 查询订单所属用户
	var order Order
	db.Preload("User").First(&order, 1)

	if order.User.ID == 0 {
		t.Error("订单所属用户的 ID 不应为 0")
	}
	if order.User.Name == "" {
		t.Error("订单所属用户的名称不应为空")
	}
}

// TestPreload 测试预加载功能
func TestPreload(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	// 创建一些订单
	var user User
	db.First(&user)
	db.Create(&Order{UserID: user.ID, Product: "商品 A", Price: 100, Quantity: 1, Status: "paid"})

	t.Run("Preload 所有订单", func(t *testing.T) {
		var users []User
		if err := db.Preload("Orders").Find(&users).Error; err != nil {
			t.Fatalf("Preload 查询失败: %v", err)
		}
		// 至少有一个用户有订单
		hasOrders := false
		for _, u := range users {
			if len(u.Orders) > 0 {
				hasOrders = true
				break
			}
		}
		if !hasOrders {
			t.Error("Preload 后应至少有一个用户包含订单数据")
		}
	})

	t.Run("Preload 带条件", func(t *testing.T) {
		var users []User
		db.Preload("Orders", "status = ?", "paid").Find(&users)
		// 至少有一个用户有 paid 订单
		for _, u := range users {
			for _, o := range u.Orders {
				if o.Status != "paid" {
					t.Errorf("Preload 带条件后应只包含 paid 订单，实际得到 %s", o.Status)
				}
			}
		}
	})
}

// TestTransaction 测试事务功能
func TestTransaction(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	t.Run("成功的事务", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			newOrder := Order{
				UserID:   1,
				Product:  "事务测试商品",
				Price:    99.00,
				Quantity: 1,
				Status:   "pending",
			}
			if err := tx.Create(&newOrder).Error; err != nil {
				return err
			}
			return tx.Model(&newOrder).Update("status", "paid").Error
		})
		if err != nil {
			t.Fatalf("事务执行失败: %v", err)
		}

		// 验证订单存在
		var order Order
		db.First(&order, 1)
		if order.Status != "paid" {
			t.Errorf("预期 order.Status=paid，实际得到 %s", order.Status)
		}
	})

	t.Run("失败的事务应回滚", func(t *testing.T) {
		beforeCount := int64(0)
		db.Model(&Order{}).Count(&beforeCount)

		// 事务中故意返回错误
		err := db.Transaction(func(tx *gorm.DB) error {
			tx.Create(&Order{UserID: 1, Product: "回滚测试", Price: 10, Quantity: 1})
			return gorm.ErrInvalidData // 返回错误，应回滚
		})
		if err == nil {
			t.Fatal("事务应返回错误")
		}

		// 订单数应不变
		afterCount := int64(0)
		db.Model(&Order{}).Count(&afterCount)
		if beforeCount != afterCount {
			t.Errorf("回滚后订单数应不变 (%d)，实际为 %d", beforeCount, afterCount)
		}
	})

	t.Run("嵌套事务", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			tx.Create(&User{Name: "嵌套用户", Email: "nested@example.com", Age: 20})

			tx.Transaction(func(tx2 *gorm.DB) error {
				tx2.Create(&Order{UserID: 1, Product: "嵌套订单", Price: 50, Quantity: 1})
				return nil
			})
			return nil
		})
		if err != nil {
			t.Fatalf("嵌套事务执行失败: %v", err)
		}

		// 验证嵌套用户和订单都已创建
		var user User
		err = db.Where("email = ?", "nested@example.com").First(&user).Error
		if err != nil {
			t.Errorf("嵌套事务中创建的用户应存在: %v", err)
		}
	})
}

// TestRawSQL 测试原生 SQL
func TestRawSQL(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)

	// 创建测试订单
	db.Create(&Order{UserID: 1, Product: "商品 A", Price: 100, Quantity: 2, Status: "paid"})
	db.Create(&Order{UserID: 1, Product: "商品 B", Price: 50, Quantity: 1, Status: "pending"})

	t.Run("Raw 查询", func(t *testing.T) {
		type Stats struct {
			Status       string
			OrderCount   int
			TotalRevenue float64
		}
		var stats []Stats
		result := db.Raw(`
			SELECT status, COUNT(*) AS order_count, SUM(price * quantity) AS total_revenue
			FROM orders GROUP BY status ORDER BY status
		`).Scan(&stats)

		if result.Error != nil {
			t.Fatalf("Raw 查询失败: %v", result.Error)
		}
		if len(stats) == 0 {
			t.Fatal("预期有统计数据")
		}
	})

	t.Run("Exec 执行原生更新", func(t *testing.T) {
		result := db.Exec("UPDATE orders SET status = ? WHERE status = ?", "cancelled", "pending")
		if result.Error != nil {
			t.Fatalf("Exec 失败: %v", result.Error)
		}

		var count int64
		db.Model(&Order{}).Where("status = ?", "cancelled").Count(&count)
		if count != 1 {
			t.Errorf("预期 1 个 cancelled 订单，实际得到 %d", count)
		}
	})

	t.Run("命名参数", func(t *testing.T) {
		var count int64
		result := db.Raw("SELECT COUNT(*) FROM orders WHERE status = @status", map[string]interface{}{
			"status": "paid",
		}).Scan(&count)

		if result.Error != nil {
			t.Fatalf("命名参数查询失败: %v", result.Error)
		}
		if count != 1 {
			t.Errorf("预期 1 个 paid 订单，实际得到 %d", count)
		}
	})
}

// TestGormExpr 测试 GORM 表达式
func TestGormExpr(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	db.Create(&Order{UserID: 1, Product: "测试商品", Price: 100, Quantity: 1, Status: "paid"})

	// 使用表达式更新价格
	result := db.Model(&Order{}).Where("id = ?", 1).
		Update("price", gorm.Expr("price * ?", 1.1))
	if result.Error != nil {
		t.Fatalf("Expr 更新失败: %v", result.Error)
	}

	var order Order
	db.First(&order, 1)
	expected := 110.0
	if order.Price < expected-0.01 || order.Price > expected+0.01 {
		t.Errorf("预期价格 %.2f，实际得到 %.2f", expected, order.Price)
	}
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	db := setupTestDB(t)

	// 批量创建
	batchUsers := []User{
		{Name: "用户A", Email: "a@example.com", Age: 20},
		{Name: "用户B", Email: "b@example.com", Age: 25},
		{Name: "用户C", Email: "c@example.com", Age: 30},
	}
	result := db.Create(&batchUsers)
	if result.Error != nil {
		t.Fatalf("批量创建失败: %v", result.Error)
	}
	if result.RowsAffected != 3 {
		t.Errorf("预期影响 3 行，实际影响 %d", result.RowsAffected)
	}

	// 批量查询年龄范围
	var users []User
	db.Where("age BETWEEN ? AND ?", 20, 30).Find(&users)
	if len(users) != 3 {
		t.Errorf("预期 3 个用户，实际得到 %d", len(users))
	}
}