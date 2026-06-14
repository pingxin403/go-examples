// main.go - GORM ORM 示例
//
// 本示例演示 GORM 的核心功能：
//   - 定义模型 (User, Order)
//   - AutoMigrate 自动迁移
//   - CRUD 操作
//   - Has Many / Belongs To 关联关系
//   - Preload / Joins 预加载
//   - 事务
//   - 原生 SQL
//
// 使用 SQLite 内存数据库，无需外部依赖即可运行。
//
// 安装依赖:
//   go get gorm.io/gorm gorm.io/driver/sqlite

package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// =========================================================
// 1. 模型定义
// =========================================================

// User 用户模型 — 与 Order 存在 Has Many 关系
type User struct {
	ID        uint           `gorm:"primarykey"`
	Name      string         `gorm:"type:varchar(100);not null;index"` // 用户名，建索引
	Email     string         `gorm:"type:varchar(200);uniqueIndex"`     // 邮箱，唯一索引
	Age       int            `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除字段

	// Has Many 关系：一个用户有多笔订单
	Orders []Order `gorm:"foreignKey:UserID"`
}

// Order 订单模型 — 与 User 存在 Belongs To 关系
type Order struct {
	ID        uint    `gorm:"primarykey"`
	UserID    uint    `gorm:"not null;index"` // 外键
	Product   string  `gorm:"type:varchar(200);not null"`
	Price     float64 `gorm:"type:decimal(10,2);not null"`
	Quantity  int     `gorm:"not null;default:1"`
	Status    string  `gorm:"type:varchar(20);default:'pending'"` // pending / paid / shipped / cancelled
	CreatedAt time.Time
	UpdatedAt time.Time

	// Belongs To 关系：订单属于某个用户
	User User `gorm:"foreignKey:UserID"`
}

// TableName 自定义 Order 表名
func (Order) TableName() string {
	return "orders"
}

// =========================================================
// 2. 辅助函数
// =========================================================

// mustExec 执行操作并检查错误
func mustExec(label string, fn func() error) {
	if err := fn(); err != nil {
		log.Fatalf("[%s] 失败: %v", label, err)
	}
	log.Printf("[%s] 成功", label)
}

// logQuery 打印查询结果标题
func logQuery(title string) {
	fmt.Printf("\n========== %s ==========\n", title)
}

// =========================================================
// 3. 主函数
// =========================================================

func main() {
	// =========================================================
	// 3.1 连接 SQLite 内存数据库
	// =========================================================
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // 只打印 Warn 级别日志，保持输出简洁
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	log.Println("已连接 SQLite 内存数据库")

	// =========================================================
	// 3.2 AutoMigrate — 自动创建/更新表结构
	// =========================================================
	mustExec("AutoMigrate", func() error {
		return db.AutoMigrate(&User{}, &Order{})
	})
	log.Println("表结构自动迁移完成")

	// =========================================================
	// 3.3 CRUD 操作
	// =========================================================
	basicCRUD(db)

	// =========================================================
	// 3.4 关联操作 — Has Many / Belongs To
	// =========================================================
	relationDemo(db)

	// =========================================================
	// 3.5 预加载 — Preload / Joins
	// =========================================================
	eagerLoadingDemo(db)

	// =========================================================
	// 3.6 事务演示
	// =========================================================
	transactionDemo(db)

	// =========================================================
	// 3.7 原生 SQL
	// =========================================================
	rawSQLDemo(db)

	log.Println("\n所有示例执行完毕！")
}

// =========================================================
// 4. 基本 CRUD 操作
// =========================================================

func basicCRUD(db *gorm.DB) {
	// ---------- Create ----------
	logQuery("CREATE — 创建记录")

	users := []User{
		{Name: "张三", Email: "zhangsan@example.com", Age: 28},
		{Name: "李四", Email: "lisi@example.com", Age: 35},
		{Name: "王五", Email: "wangwu@example.com", Age: 22},
	}

	for _, u := range users {
		result := db.Create(&u)
		if result.Error != nil {
			log.Printf("创建用户失败: %v", result.Error)
		} else {
			log.Printf("创建用户: ID=%d, Name=%s", u.ID, u.Name)
		}
	}

	// 批量创建
	batchUsers := []User{
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 30},
		{Name: "钱七", Email: "qianqi@example.com", Age: 25},
	}
	result := db.Create(&batchUsers)
	log.Printf("批量创建: 影响了 %d 行", result.RowsAffected)

	// ---------- Read ----------
	logQuery("READ — 查询记录")

	// 查询单条记录
	var user User
	db.First(&user, 1) // 按主键查询
	log.Printf("First ID=1: Name=%s, Email=%s", user.Name, user.Email)

	// 按条件查询
	var usersFound []User
	db.Where("age > ?", 25).Find(&usersFound)
	log.Printf("年龄 > 25 的用户: %d 个", len(usersFound))
	for _, u := range usersFound {
		log.Printf("  - %s (%d岁)", u.Name, u.Age)
	}

	// 使用 Struct 条件查询
	db.Where(&User{Name: "李四"}).First(&user)
	log.Printf("查询 Name=李四: Email=%s", user.Email)

	// 使用 Map 条件查询
	db.Where(map[string]interface{}{"age": 22}).Find(&usersFound)
	log.Printf("查询 age=22 的用户: %d 个", len(usersFound))

	// ---------- Update ----------
	logQuery("UPDATE — 更新记录")

	// 更新单个字段
	db.Model(&User{}).Where("name = ?", "王五").Update("age", 23)
	log.Println("王五的年龄更新为 23")

	// 更新多个字段
	db.Model(&User{}).Where("name = ?", "张三").Updates(User{Name: "张三丰", Age: 29})
	log.Println("张三更新为张三丰, age=29")

	// ---------- Delete ----------
	logQuery("DELETE — 删除记录（软删除）")

	// 软删除（User 模型有 DeletedAt 字段）
	db.Where("name = ?", "钱七").Delete(&User{})
	log.Println("软删除用户: 钱七")

	// 查询全部（包含软删除的）
	var allUsers []User
	db.Unscoped().Find(&allUsers)
	log.Printf("全部用户（含已删除）: %d 个", len(allUsers))

	// 查询未删除的
	db.Find(&allUsers)
	log.Printf("未删除用户: %d 个", len(allUsers))
}

// =========================================================
// 5. 关联操作演示
// =========================================================

func relationDemo(db *gorm.DB) {
	logQuery("关联 — Has Many / Belongs To")

	// 获取第一个用户
	var user User
	db.First(&user)

	// 创建该用户的订单
	orders := []Order{
		{UserID: user.ID, Product: "Go 语言圣经", Price: 89.00, Quantity: 1, Status: "paid"},
		{UserID: user.ID, Product: "机械键盘", Price: 399.00, Quantity: 2, Status: "pending"},
		{UserID: user.ID, Product: "显示器支架", Price: 159.00, Quantity: 1, Status: "shipped"},
	}

	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			log.Printf("创建订单失败: %v", err)
		} else {
			log.Printf("创建订单: ID=%d, Product=%s, Price=%.2f", o.ID, o.Product, o.Price)
		}
	}

	// 关联查询：通过 User 查询其所有 Orders
	var userWithOrders User
	db.Preload("Orders").First(&userWithOrders, user.ID)
	log.Printf("用户 %s 的订单:", userWithOrders.Name)
	for _, o := range userWithOrders.Orders {
		log.Printf("  - #%d %s ¥%.2f x%d [%s]", o.ID, o.Product, o.Price, o.Quantity, o.Status)
	}

	// 反向关联：通过 Order 查询所属 User
	var order Order
	db.Preload("User").First(&order, 1)
	log.Printf("订单 #%d 属于用户: %s (%s)", order.ID, order.User.Name, order.User.Email)
}

// =========================================================
// 6. 预加载（Eager Loading）演示
// =========================================================

func eagerLoadingDemo(db *gorm.DB) {
	logQuery("预加载 — Preload / Joins")

	// ---------- Preload ----------
	// Preload: 先查 User 表，再分别查每个 User 的 Orders（N+1 问题被 GORM 优化为 2 次查询）
	var users []User
	db.Preload("Orders").Find(&users)
	log.Println("Preload 方式加载所有用户及其订单:")
	for _, u := range users {
		log.Printf("  %s: %d 笔订单", u.Name, len(u.Orders))
	}

	// ---------- Preload 带条件 ----------
	var paidOrdersUsers []User
	db.Preload("Orders", "status = ?", "paid").Find(&paidOrdersUsers)
	log.Println("仅预加载已支付的订单:")
	for _, u := range paidOrdersUsers {
		for _, o := range u.Orders {
			log.Printf("  %s: #%d %s [%s]", u.Name, o.ID, o.Product, o.Status)
		}
	}

	// ---------- Joins ----------
	// Joins: 使用 JOIN 一次性查出 User + Order（适合需要过滤关联表数据的场景）
	type UserOrderSummary struct {
		UserID   uint
		UserName string
		Total    float64
	}

	var summaries []UserOrderSummary
	db.Model(&Order{}).
		Select("orders.user_id, users.name as user_name, SUM(orders.price * orders.quantity) as total").
		Joins("left join users on users.id = orders.user_id").
		Group("orders.user_id").
		Scan(&summaries)

	log.Println("用户消费总额（Joins 查询）:")
	for _, s := range summaries {
		log.Printf("  %s: ¥%.2f", s.UserName, s.Total)
	}
}

// =========================================================
// 7. 事务演示
// =========================================================

func transactionDemo(db *gorm.DB) {
	logQuery("事务 — 确保原子操作")

	// 成功的事务：创建订单 + 更新订单状态
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建新订单
		newOrder := Order{
			UserID:   1,
			Product:  "编程书籍套装",
			Price:    199.00,
			Quantity: 1,
			Status:   "pending",
		}
		if err := tx.Create(&newOrder).Error; err != nil {
			return err // 发生错误，回滚事务
		}
		log.Printf("[事务] 创建订单 #%d", newOrder.ID)

		// 2. 更新订单状态为 paid
		if err := tx.Model(&newOrder).Update("status", "paid").Error; err != nil {
			return err
		}
		log.Printf("[事务] 订单 #%d 状态更新为 paid", newOrder.ID)

		return nil // 提交事务
	})

	if err != nil {
		log.Fatalf("[事务] 失败: %v", err)
	}
	log.Println("[事务] 成功提交")

	// ---------- 嵌套事务 ----------
	logQuery("嵌套事务")

	err = db.Transaction(func(tx *gorm.DB) error {
		// 外层事务
		tx.Create(&User{Name: "嵌套用户", Email: "nested@example.com", Age: 20})

		// 嵌套事务中的 SavePoint
		tx.Transaction(func(tx2 *gorm.DB) error {
			tx2.Create(&Order{UserID: 1, Product: "嵌套订单商品", Price: 50.00, Quantity: 1})
			// 这里返回错误会回滚到 SavePoint，但外层仍可继续
			return nil
		})

		return nil
	})

	log.Printf("嵌套事务执行结果: %v", err)
}

// =========================================================
// 8. 原生 SQL 演示
// =========================================================

func rawSQLDemo(db *gorm.DB) {
	logQuery("原生 SQL — Raw / Exec")

	// ---------- Raw 查询 ----------
	type Stats struct {
		Status       string
		OrderCount   int
		TotalRevenue float64
	}

	var stats []Stats
	db.Raw(`
		SELECT
			status,
			COUNT(*)      AS order_count,
			SUM(price * quantity) AS total_revenue
		FROM orders
		GROUP BY status
		ORDER BY status
	`).Scan(&stats)

	log.Println("各状态订单统计（原生 SQL）:")
	for _, s := range stats {
		log.Printf("  %s: %d 笔, 总额 ¥%.2f", s.Status, s.OrderCount, s.TotalRevenue)
	}

	// ---------- Exec 执行 ----------
	// 使用原生 SQL 更新
	db.Exec("UPDATE orders SET status = ? WHERE status = ?", "cancelled", "pending")
	log.Println("将所有 pending 状态的订单更新为 cancelled")

	// 使用命名参数
	var count int64
	db.Raw("SELECT COUNT(*) FROM orders WHERE status = @status", map[string]interface{}{
		"status": "cancelled",
	}).Scan(&count)
	log.Printf("已取消的订单数量: %d", count)

	// ---------- 使用 clause 构建复杂查询 ----------
	logQuery("Clause 构建器")

	var expensiveOrders []Order
	db.Clauses(clause.OrderBy{
		Expression: clause.Expr{SQL: "price * quantity DESC", Vars: []interface{}{}, WithoutParentheses: true},
	}).Limit(3).Find(&expensiveOrders)

	log.Println("金额最高的 3 笔订单:")
	for _, o := range expensiveOrders {
		log.Printf("  #%d %s ¥%.2f x%d = ¥%.2f", o.ID, o.Product, o.Price, o.Quantity, o.Price*float64(o.Quantity))
	}

	// ---------- 使用表达式 ----------
	logQuery("GORM 表达式")

	// 使用 gorm.Expr 做字段级计算
	db.Model(&Order{}).Where("id = ?", 1).
		Update("price", gorm.Expr("price * ?", 1.1))
	log.Println("订单 #1 的价格上调 10%")
}