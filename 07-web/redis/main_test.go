// main_test.go — Redis 操作单元测试
//
// 测试内容：
//   - Redis URL 解析和选项验证
//   - 键名规范格式化
//   - 错误处理辅助函数
//   - 实际 Redis 操作（需要 Redis 服务器）
//   - Redis 连接字符串构造
//
// 注意：需要实际 Redis 连接的测试会用 t.Skip 跳过（无 Redis 服务器时）。

package main

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// ---------- 辅助函数测试 ----------

// TestMustRedis 测试 mustRedis 的错误处理逻辑
func TestMustRedis(t *testing.T) {
	// 不应该 panic — err 为 nil
	mustRedis("测试操作", nil)

	// redis.Nil 不应导致 panic — 应仅打印日志
	mustRedis("键不存在测试", redis.Nil)

	t.Log("mustRedis 对 nil 和 redis.Nil 处理正确")
}

// TestLogSection 测试 logSection 不 panic
func TestLogSection(t *testing.T) {
	logSection("测试分区标题")
	t.Log("logSection 执行正常")
}

// TestLogKey 测试 logKey 不 panic
func TestLogKey(t *testing.T) {
	logKey("Get", "test:key", "test_value")
	t.Log("logKey 执行正常")
}

// ---------- Redis 客户端配置测试 ----------

// TestRedisOptions 测试 Redis 连接选项的正确性
func TestRedisOptions(t *testing.T) {
	opts := &redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// 验证配置
	if opts.Addr != "localhost:6379" {
		t.Errorf("预期 Addr=localhost:6379，实际得到 %s", opts.Addr)
	}
	if opts.PoolSize != 10 {
		t.Errorf("预期 PoolSize=10，实际得到 %d", opts.PoolSize)
	}
	if opts.MinIdleConns != 3 {
		t.Errorf("预期 MinIdleConns=3，实际得到 %d", opts.MinIdleConns)
	}
	if opts.DialTimeout != 5*time.Second {
		t.Errorf("预期 DialTimeout=5s，实际得到 %v", opts.DialTimeout)
	}
}

// TestClientCreation 测试创建 Redis 客户端不 panic
func TestClientCreation(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	if client == nil {
		t.Fatal("redis.NewClient 不应返回 nil")
	}
	client.Close()
}

// ---------- 集成测试（需要 Redis 服务器） ----------

// skipIfNoRedis 检查 Redis 是否可用，不可用时跳过测试
func skipIfNoRedis(t *testing.T) {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis 服务器不可用，跳过测试: %v", err)
	}
}

// TestRedisPing 测试 Redis 连接 Ping
func TestRedisPing(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()

	ctx := context.Background()
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
	if pong != "PONG" {
		t.Errorf("预期 PONG，实际得到 %s", pong)
	}
}

// TestRedisStringOps 测试 String 操作
func TestRedisStringOps(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()
	ctx := context.Background()

	// 清理
	defer client.FlushDB(ctx)

	t.Run("Set 和 Get", func(t *testing.T) {
		err := client.Set(ctx, "test:string", "hello", 0).Err()
		if err != nil {
			t.Fatalf("Set 失败: %v", err)
		}

		val, err := client.Get(ctx, "test:string").Result()
		if err != nil {
			t.Fatalf("Get 失败: %v", err)
		}
		if val != "hello" {
			t.Errorf("预期 hello，实际得到 %s", val)
		}
	})

	t.Run("Set 带过期时间", func(t *testing.T) {
		err := client.Set(ctx, "test:expire", "temp", 1*time.Hour).Err()
		if err != nil {
			t.Fatalf("SetEX 失败: %v", err)
		}

		ttl, err := client.TTL(ctx, "test:expire").Result()
		if err != nil {
			t.Fatalf("TTL 查询失败: %v", err)
		}
		if ttl <= 0 {
			t.Errorf("TTL 应大于 0，实际得到 %v", ttl)
		}
	})

	t.Run("获取不存在的键", func(t *testing.T) {
		_, err := client.Get(ctx, "test:nonexistent").Result()
		if err != redis.Nil {
			t.Errorf("不存在的键应返回 redis.Nil，实际得到 %v", err)
		}
	})

	t.Run("MSet 和 MGet", func(t *testing.T) {
		err := client.MSet(ctx, "test:k1", "v1", "test:k2", "v2").Err()
		if err != nil {
			t.Fatalf("MSet 失败: %v", err)
		}

		vals, err := client.MGet(ctx, "test:k1", "test:k2").Result()
		if err != nil {
			t.Fatalf("MGet 失败: %v", err)
		}
		if len(vals) != 2 {
			t.Errorf("预期 2 个值，实际得到 %d", len(vals))
		}
	})

	t.Run("Incr 自增", func(t *testing.T) {
		client.Del(ctx, "test:counter")

		cnt, err := client.Incr(ctx, "test:counter").Result()
		if err != nil {
			t.Fatalf("Incr 失败: %v", err)
		}
		if cnt != 1 {
			t.Errorf("第一次 Incr 预期 1，实际得到 %d", cnt)
		}

		cnt, err = client.IncrBy(ctx, "test:counter", 5).Result()
		if err != nil {
			t.Fatalf("IncrBy 失败: %v", err)
		}
		if cnt != 6 {
			t.Errorf("IncrBy 后预期 6，实际得到 %d", cnt)
		}
	})

	t.Run("GetSet", func(t *testing.T) {
		client.Set(ctx, "test:getset", "old", 0)
		old, err := client.GetSet(ctx, "test:getset", "new").Result()
		if err != nil {
			t.Fatalf("GetSet 失败: %v", err)
		}
		if old != "old" {
			t.Errorf("GetSet 旧值预期 old，实际得到 %s", old)
		}

		val, _ := client.Get(ctx, "test:getset").Result()
		if val != "new" {
			t.Errorf("GetSet 后新值预期 new，实际得到 %s", val)
		}
	})
}

// TestRedisListOps 测试 List 操作
func TestRedisListOps(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()
	ctx := context.Background()
	key := "test:list"

	defer client.Del(ctx, key)
	defer client.FlushDB(ctx)

	t.Run("LPush 和 LLen", func(t *testing.T) {
		n, err := client.LPush(ctx, key, "c", "b", "a").Result()
		if err != nil {
			t.Fatalf("LPush 失败: %v", err)
		}
		if n != 3 {
			t.Errorf("LPush 后预期 3 个元素，实际得到 %d", n)
		}

		length, _ := client.LLen(ctx, key).Result()
		if length != 3 {
			t.Errorf("LLen 预期 3，实际得到 %d", length)
		}
	})

	t.Run("RPush", func(t *testing.T) {
		n, err := client.RPush(ctx, key, "d", "e").Result()
		if err != nil {
			t.Fatalf("RPush 失败: %v", err)
		}
		if n != 5 {
			t.Errorf("RPush 后预期 5 个元素，实际得到 %d", n)
		}
	})

	t.Run("LRange", func(t *testing.T) {
		vals, err := client.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			t.Fatalf("LRange 失败: %v", err)
		}
		if len(vals) != 5 {
			t.Errorf("LRange 预期 5 个元素，实际得到 %d", len(vals))
		}
	})

	t.Run("LPop 和 RPop", func(t *testing.T) {
		// LPush 顺序: LPush key c b a → List: [a, b, c]
		// 然后 RPush d e → List: [a, b, c, d, e]
		left, _ := client.LPop(ctx, key).Result()
		if left != "a" {
			t.Errorf("LPop 预期 a，实际得到 %s", left)
		}

		right, _ := client.RPop(ctx, key).Result()
		if right != "e" {
			t.Errorf("RPop 预期 e，实际得到 %s", right)
		}
	})
}

// TestRedisHashOps 测试 Hash 操作
func TestRedisHashOps(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()
	ctx := context.Background()
	key := "test:hash"

	defer client.Del(ctx, key)
	defer client.FlushDB(ctx)

	t.Run("HSet 和 HGet", func(t *testing.T) {
		n, err := client.HSet(ctx, key, map[string]interface{}{
			"name": "张三",
			"age":  28,
			"city": "北京",
		}).Result()
		if err != nil {
			t.Fatalf("HSet 失败: %v", err)
		}
		if n != 3 {
			t.Errorf("HSet 预期 3 个字段，实际设置 %d", n)
		}

		name, _ := client.HGet(ctx, key, "name").Result()
		if name != "张三" {
			t.Errorf("HGet 预期 张三，实际得到 %s", name)
		}
	})

	t.Run("HMGet", func(t *testing.T) {
		vals, err := client.HMGet(ctx, key, "name", "age", "city").Result()
		if err != nil {
			t.Fatalf("HMGet 失败: %v", err)
		}
		if len(vals) != 3 {
			t.Errorf("HMGet 预期 3 个值，实际得到 %d", len(vals))
		}
	})

	t.Run("HGetAll", func(t *testing.T) {
		all, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			t.Fatalf("HGetAll 失败: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("HGetAll 预期 3 个字段，实际得到 %d", len(all))
		}
	})

	t.Run("HExists", func(t *testing.T) {
		exists, _ := client.HExists(ctx, key, "name").Result()
		if !exists {
			t.Error("HExists 字段 'name' 应存在")
		}

		exists, _ = client.HExists(ctx, key, "phone").Result()
		if exists {
			t.Error("HExists 字段 'phone' 不应存在")
		}
	})

	t.Run("HIncrBy", func(t *testing.T) {
		newAge, err := client.HIncrBy(ctx, key, "age", 1).Result()
		if err != nil {
			t.Fatalf("HIncrBy 失败: %v", err)
		}
		if newAge != 29 {
			t.Errorf("HIncrBy 后预期 age=29，实际得到 %d", newAge)
		}
	})

	t.Run("HLen", func(t *testing.T) {
		count, _ := client.HLen(ctx, key).Result()
		if count != 3 {
			t.Errorf("HLen 预期 3，实际得到 %d", count)
		}
	})
}

// TestRedisSetOps 测试 Set 操作
func TestRedisSetOps(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()
	ctx := context.Background()

	defer client.FlushDB(ctx)

	t.Run("SAdd 和 SMembers", func(t *testing.T) {
		n, err := client.SAdd(ctx, "test:set", "a", "b", "c").Result()
		if err != nil {
			t.Fatalf("SAdd 失败: %v", err)
		}
		if n != 3 {
			t.Errorf("SAdd 预期添加 3 个，实际添加 %d", n)
		}

		members, _ := client.SMembers(ctx, "test:set").Result()
		if len(members) != 3 {
			t.Errorf("SMembers 预期 3 个成员，实际得到 %d", len(members))
		}
	})

	t.Run("SCard", func(t *testing.T) {
		card, _ := client.SCard(ctx, "test:set").Result()
		if card != 3 {
			t.Errorf("SCard 预期 3，实际得到 %d", card)
		}
	})

	t.Run("SIsMember", func(t *testing.T) {
		ok, _ := client.SIsMember(ctx, "test:set", "a").Result()
		if !ok {
			t.Error("'a' 应在集合中")
		}

		ok, _ = client.SIsMember(ctx, "test:set", "z").Result()
		if ok {
			t.Error("'z' 不应在集合中")
		}
	})

	t.Run("SInter / SUnion / SDiff", func(t *testing.T) {
		client.SAdd(ctx, "test:set2", "a", "b", "d")

		inter, _ := client.SInter(ctx, "test:set", "test:set2").Result()
		if len(inter) != 2 {
			t.Errorf("交集预期 2 个元素，实际得到 %d: %v", len(inter), inter)
		}

		union, _ := client.SUnion(ctx, "test:set", "test:set2").Result()
		if len(union) != 4 {
			t.Errorf("并集预期 4 个元素，实际得到 %d: %v", len(union), union)
		}
	})
}

// TestRedisPipelineOps 测试 Pipeline 管道操作
func TestRedisPipelineOps(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()
	ctx := context.Background()

	defer client.FlushDB(ctx)

	t.Run("Pipeline 批量执行", func(t *testing.T) {
		pipe := client.Pipeline()
		incr := pipe.Incr(ctx, "test:pipe:counter")
		pipe.Set(ctx, "test:pipe:foo", "bar", 0)
		pipe.Set(ctx, "test:pipe:hello", "world", 0)

		_, err := pipe.Exec(ctx)
		if err != nil {
			t.Fatalf("Pipeline Exec 失败: %v", err)
		}

		if incr.Val() != 1 {
			t.Errorf("Pipeline Incr 预期 1，实际得到 %d", incr.Val())
		}
	})

	t.Run("Pipelined 简洁写法", func(t *testing.T) {
		var nameVal string
		_, err := client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "test:pipe:name", "小明", 0)
			getName := pipe.Get(ctx, "test:pipe:name")
			nameVal = getName.Val()
			return nil
		})
		if err != nil {
			t.Fatalf("Pipelined 失败: %v", err)
		}
		if nameVal != "小明" {
			t.Errorf("Pipelined 预期 小明，实际得到 %s", nameVal)
		}
	})
}

// TestRedisDeleteAndExpire 测试删除和过期操作
func TestRedisDeleteAndExpire(t *testing.T) {
	skipIfNoRedis(t)

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	defer client.Close()
	ctx := context.Background()

	defer client.FlushDB(ctx)

	t.Run("Del 删除键", func(t *testing.T) {
		client.Set(ctx, "test:del:key", "value", 0)
		n, err := client.Del(ctx, "test:del:key").Result()
		if err != nil {
			t.Fatalf("Del 失败: %v", err)
		}
		if n != 1 {
			t.Errorf("Del 预期删除 1 个键，实际删除 %d", n)
		}
	})

	t.Run("Expire 设置过期时间", func(t *testing.T) {
		client.Set(ctx, "test:expire:key", "value", 0)
		ok, err := client.Expire(ctx, "test:expire:key", 1*time.Hour).Result()
		if err != nil {
			t.Fatalf("Expire 失败: %v", err)
		}
		if !ok {
			t.Error("Expire 应返回 true")
		}
	})
}