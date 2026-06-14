// main.go - Go Redis 客户端示例
//
// 本示例演示 go-redis 的核心功能：
//   - 连接 Redis
//   - String 操作 (Set, Get, MSet, MGet)
//   - List 操作 (LPush, RPop, LLen)
//   - Hash 操作 (HSet, HGet, HGetAll)
//   - Set 操作 (SAdd, SMembers)
//   - Pipeline 管道
//   - Pub/Sub 发布订阅
//
// 安装依赖:
//   go get github.com/redis/go-redis/v9
//
// 前置条件：需要运行 Redis 服务器
//   docker run --name redis-demo -p 6379:6379 -d redis:7-alpine

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ---------- 全局变量 ----------

var (
	ctx = context.Background()
	rdb *redis.Client
)

// ---------- 辅助函数 ----------

// logSection 打印分区标题
func logSection(title string) {
	fmt.Printf("\n═══════════ %s ═══════════\n", title)
}

// logKey 打印操作结果
func logKey(operation, key, value string) {
	log.Printf("[%s] %s = %s", operation, key, value)
}

// mustRedis 检查 Redis 操作错误
func mustRedis(label string, err error) {
	if err != nil {
		if err == redis.Nil {
			log.Printf("[%s] 键不存在", label)
			return
		}
		log.Fatalf("[%s] 失败: %v", label, err)
	}
}

// =========================================================
// 1. 连接 Redis
// =========================================================

func connectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379", // Redis 地址
		Password:     "",               // 密码（无密码则为空）
		DB:           0,                // 使用默认数据库
		PoolSize:     10,               // 连接池大小
		MinIdleConns: 3,                // 最小空闲连接数
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读取超时
		WriteTimeout: 3 * time.Second,  // 写入超时
	})

	// 检查连接是否正常
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("连接 Redis 失败: %v\n请确保 Redis 已启动: docker run --name redis-demo -p 6379:6379 -d redis:7-alpine", err)
	}
	log.Printf("Redis 连接成功: %s", pong)

	return client
}

// =========================================================
// 2. String 操作
// =========================================================

func stringOperations() {
	logSection("String 操作")

	// Set: 设置键值对
	err := rdb.Set(ctx, "user:1:name", "张三", 0).Err()
	mustRedis("Set", err)
	logKey("Set", "user:1:name", "张三")

	// Set 带过期时间（10 秒后自动删除）
	err = rdb.Set(ctx, "temp:token", "abc123", 10*time.Second).Err()
	mustRedis("SetEX", err)
	log.Println("[SetEX] temp:token = abc123 (10 秒后过期)")

	// Get: 获取值
	val, err := rdb.Get(ctx, "user:1:name").Result()
	mustRedis("Get", err)
	logKey("Get", "user:1:name", val)

	// 获取不存在的键
	_, err = rdb.Get(ctx, "nonexistent").Result()
	if err == redis.Nil {
		log.Println("[Get] nonexistent: 键不存在（返回 redis.Nil）")
	}

	// MSet: 批量设置
	err = rdb.MSet(ctx, "user:2:name", "李四", "user:2:email", "lisi@example.com").Err()
	mustRedis("MSet", err)
	log.Println("[MSet] 批量设置 user:2 信息")

	// MGet: 批量获取
	vals, err := rdb.MGet(ctx, "user:1:name", "user:2:name", "user:2:email").Result()
	mustRedis("MGet", err)
	log.Printf("[MGet] 批量获取结果: %v", vals)

	// Incr: 自增（适合做计数器）
	cnt, err := rdb.Incr(ctx, "counter:page_views").Result()
	mustRedis("Incr", err)
	log.Printf("[Incr] counter:page_views = %d", cnt)

	// IncrBy: 指定步长自增
	cnt, err = rdb.IncrBy(ctx, "counter:page_views", 5).Result()
	mustRedis("IncrBy", err)
	log.Printf("[IncrBy] counter:page_views = %d", cnt)

	// GetSet: 获取旧值并设置新值
	oldVal, err := rdb.GetSet(ctx, "user:1:name", "张三丰").Result()
	mustRedis("GetSet", err)
	log.Printf("[GetSet] 旧值=%s, 新值=%s", oldVal, "张三丰")
}

// =========================================================
// 3. List 操作
// =========================================================

func listOperations() {
	logSection("List 操作")

	key := "list:messages"

	// 清空已有数据
	rdb.Del(ctx, key)

	// LPush: 从左侧推入（栈行为）
	err := rdb.LPush(ctx, key, "消息3", "消息2", "消息1").Err()
	mustRedis("LPush", err)
	log.Println("[LPush] 推入 3 条消息")

	// RPush: 从右侧推入（队列行为）
	err = rdb.RPush(ctx, key, "消息4", "消息5").Err()
	mustRedis("RPush", err)
	log.Println("[RPush] 从右侧推入 2 条消息")

	// LLen: 获取列表长度
	length, err := rdb.LLen(ctx, key).Result()
	mustRedis("LLen", err)
	log.Printf("[LLen] 列表长度: %d", length)

	// LRange: 获取范围元素
	all, err := rdb.LRange(ctx, key, 0, -1).Result()
	mustRedis("LRange", err)
	log.Printf("[LRange] 所有元素: %v", all)

	// LPop: 从左侧弹出
	msg, err := rdb.LPop(ctx, key).Result()
	mustRedis("LPop", err)
	log.Printf("[LPop] 弹出: %s", msg)

	// RPop: 从右侧弹出
	msg, err = rdb.RPop(ctx, key).Result()
	mustRedis("RPop", err)
	log.Printf("[RPop] 弹出: %s", msg)

	// BLPop: 阻塞式弹出（带超时）
	// 如果列表为空，在超时时间内阻塞等待
	result, err := rdb.BLPop(ctx, 2*time.Second, key).Result()
	if err != nil {
		log.Printf("[BLPop] 超时（2 秒内无新元素）: %v", err)
	} else {
		log.Printf("[BLPop] 弹出: %v", result)
	}

	// 查看最终列表内容
	remaining, _ := rdb.LRange(ctx, key, 0, -1).Result()
	log.Printf("最终列表内容: %v", remaining)
}

// =========================================================
// 4. Hash 操作
// =========================================================

func hashOperations() {
	logSection("Hash 操作")

	key := "hash:user:1001"

	// 清除已存在数据
	rdb.Del(ctx, key)

	// HSet: 设置 Hash 字段
	err := rdb.HSet(ctx, key, map[string]interface{}{
		"name":   "王五",
		"email":  "wangwu@example.com",
		"age":    28,
		"city":   "北京",
		"skills": "Go,Python,Redis",
	}).Err()
	mustRedis("HSet", err)
	log.Printf("[HSet] 设置用户 Hash (%s)", key)

	// HGet: 获取单个字段
	name, err := rdb.HGet(ctx, key, "name").Result()
	mustRedis("HGet", err)
	log.Printf("[HGet] name = %s", name)

	// HMGet: 获取多个字段
	fields, err := rdb.HMGet(ctx, key, "name", "email", "city").Result()
	mustRedis("HMGet", err)
	log.Printf("[HMGet] name/email/city = %v", fields)

	// HGetAll: 获取所有字段
	allFields, err := rdb.HGetAll(ctx, key).Result()
	mustRedis("HGetAll", err)
	log.Println("[HGetAll] 所有字段:")
	for field, value := range allFields {
		log.Printf("  %s = %s", field, value)
	}

	// HExists: 检查字段是否存在
	exists, err := rdb.HExists(ctx, key, "phone").Result()
	mustRedis("HExists", err)
	log.Printf("[HExists] phone 字段存在? %v", exists)

	// HIncrBy: Hash 字段自增
	newAge, err := rdb.HIncrBy(ctx, key, "age", 1).Result()
	mustRedis("HIncrBy", err)
	log.Printf("[HIncrBy] age 增加 1 后: %d", newAge)

	// HLen: 获取字段数量
	count, err := rdb.HLen(ctx, key).Result()
	mustRedis("HLen", err)
	log.Printf("[HLen] 字段数量: %d", count)

	// HKeys / HVals
	keys, _ := rdb.HKeys(ctx, key).Result()
	vals, _ := rdb.HVals(ctx, key).Result()
	log.Printf("[HKeys] 所有字段名: %v", keys)
	log.Printf("[HVals] 所有字段值: %v", vals)
}

// =========================================================
// 5. Set 操作
// =========================================================

func setOperations() {
	logSection("Set 操作")

	// SAdd: 向集合添加元素
	err := rdb.SAdd(ctx, "set:tags", "golang", "redis", "gin", "gorm", "grpc").Err()
	mustRedis("SAdd", err)
	log.Println("[SAdd] 添加标签: golang, redis, gin, gorm, grpc")

	// SMembers: 获取所有成员
	members, err := rdb.SMembers(ctx, "set:tags").Result()
	mustRedis("SMembers", err)
	log.Printf("[SMembers] 所有标签: %v", members)

	// SCard: 获取集合基数
	card, err := rdb.SCard(ctx, "set:tags").Result()
	mustRedis("SCard", err)
	log.Printf("[SCard] 标签数量: %d", card)

	// SIsMember: 检查成员是否存在
	isMember, err := rdb.SIsMember(ctx, "set:tags", "golang").Result()
	mustRedis("SIsMember", err)
	log.Printf("[SIsMember] golang 在集合中? %v", isMember)

	// 创建第二个集合
	rdb.SAdd(ctx, "set:tags2", "golang", "python", "java", "redis")

	// SInter: 交集
	inter, _ := rdb.SInter(ctx, "set:tags", "set:tags2").Result()
	log.Printf("[SInter] 交集: %v", inter)

	// SUnion: 并集
	union, _ := rdb.SUnion(ctx, "set:tags", "set:tags2").Result()
	log.Printf("[SUnion] 并集: %v", union)

	// SDiff: 差集
	diff, _ := rdb.SDiff(ctx, "set:tags", "set:tags2").Result()
	log.Printf("[SDiff] 差集 (tags - tags2): %v", diff)

	// SPop: 随机弹出
	popped, err := rdb.SPop(ctx, "set:tags").Result()
	mustRedis("SPop", err)
	log.Printf("[SPop] 随机弹出: %s", popped)
}

// =========================================================
// 6. Pipeline 管道操作
// =========================================================

func pipelineDemo() {
	logSection("Pipeline 管道")

	// Pipeline 将多个命令打包一次性发送给 Redis，减少网络往返
	pipe := rdb.Pipeline()

	// 向管道中添加命令
	incr := pipe.Incr(ctx, "pipeline:counter")
	pipe.Set(ctx, "pipeline:foo", "bar", 0)
	pipe.Set(ctx, "pipeline:hello", "world", 0)
	getFoo := pipe.Get(ctx, "pipeline:foo")

	// 一次性执行所有命令
	_, err := pipe.Exec(ctx)
	mustRedis("Pipeline Exec", err)

	log.Printf("[Pipeline] counter = %d", incr.Val())
	log.Printf("[Pipeline] foo = %s", getFoo.Val())

	// 使用 Pipelined 方法（更简洁的写法）
	var nameVal, emailVal string
	_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, "pipeline:name", "小明", 0)
		pipe.Set(ctx, "pipeline:email", "xiaoming@example.com", 0)

		getName := pipe.Get(ctx, "pipeline:name")
		getEmail := pipe.Get(ctx, "pipeline:email")

		// 注意：在 Pipelined 回调中获取结果（闭包捕获）
		nameVal = getName.Val()
		emailVal = getEmail.Val()
		return nil
	})
	mustRedis("Pipelined", err)

	log.Printf("[Pipelined] name = %s, email = %s", nameVal, emailVal)
}

// =========================================================
// 7. Pub/Sub 发布订阅
// =========================================================

func pubSubDemo() {
	logSection("Pub/Sub 发布订阅")

	var wg sync.WaitGroup

	// 启动订阅者 goroutine
	wg.Add(1)
	go subscriber(&wg)

	// 稍等订阅者就绪
	time.Sleep(500 * time.Millisecond)

	// 启动发布者
	publisher()

	wg.Wait()
}

// subscriber 订阅者：订阅频道并接收消息
func subscriber(wg *sync.WaitGroup) {
	defer wg.Done()

	// 创建新的客户端用于订阅（Pub/Sub 模式需要专用连接）
	subClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer subClient.Close()

	// 订阅频道
	pubsub := subClient.Subscribe(ctx, "channel:notifications", "channel:alerts")
	defer pubsub.Close()

	// 等待确认订阅已建立
	_, err := pubsub.Receive(ctx)
	mustRedis("Subscribe", err)
	log.Println("[Subscriber] 已订阅 channel:notifications 和 channel:alerts")

	// 创建一个带超时的 context 用于退出
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 循环接收消息，5 秒后自动退出
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			log.Printf("[Subscriber] 收到消息 | 频道: %s | 消息: %s", msg.Channel, msg.Payload)
		case <-timeoutCtx.Done():
			log.Println("[Subscriber] 超时退出")
			return
		}
	}
}

// publisher 发布者：向频道发布消息
func publisher() {
	messages := []struct {
		channel string
		msg     string
	}{
		{"channel:notifications", "你有新的好友请求"},
		{"channel:alerts", "CPU 使用率超过 90%"},
		{"channel:notifications", "新版本 v2.0 已发布"},
		{"channel:alerts", "磁盘空间不足"},
	}

	for _, m := range messages {
		count, err := rdb.Publish(ctx, m.channel, m.msg).Result()
		mustRedis("Publish", err)
		log.Printf("[Publisher] 发布到 %s: %s (收到 %d 个订阅者)", m.channel, m.msg, count)
		time.Sleep(200 * time.Millisecond)
	}
}

// =========================================================
// 主函数
// =========================================================

func main() {
	// 连接 Redis
	rdb = connectRedis()
	defer rdb.Close()

	log.Println("========================================")
	log.Println("  Go Redis 功能演示")
	log.Println("========================================")

	// 清空当前数据库（保持每次运行结果一致）
	rdb.FlushDB(ctx)

	// 依次执行各类型操作
	stringOperations()
	listOperations()
	hashOperations()
	setOperations()
	pipelineDemo()
	pubSubDemo()

	fmt.Printf("\n═══════════ 全部演示完成 ═══════════\n")
}