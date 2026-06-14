package main

import (
	"sync"
	"testing"
	"time"
)

// TestCacheSetGet 测试缓存的基本 Set 和 Get 操作
func TestCacheSetGet(t *testing.T) {
	cache := NewCache(0) // 不启动清理协程
	defer cache.Stop()

	t.Run("设置和获取字符串", func(t *testing.T) {
		cache.Set("name", "张三", 0)
		v, ok := cache.Get("name")
		if !ok {
			t.Fatal("Get('name') 期望 ok=true")
		}
		if v != "张三" {
			t.Errorf("Get('name') = %v, 期望 '张三'", v)
		}
	})

	t.Run("设置和获取整数", func(t *testing.T) {
		cache.Set("age", 30, 0)
		v, ok := cache.GetInt("age")
		if !ok {
			t.Fatal("GetInt('age') 期望 ok=true")
		}
		if v != 30 {
			t.Errorf("GetInt('age') = %d, 期望 30", v)
		}
	})

	t.Run("获取不存在的 key", func(t *testing.T) {
		_, ok := cache.Get("nonexist")
		if ok {
			t.Error("Get('nonexist') 期望 ok=false")
		}
	})

	t.Run("获取 string 类型便捷方法", func(t *testing.T) {
		cache.Set("greeting", "你好", 0)
		s, ok := cache.GetString("greeting")
		if !ok {
			t.Fatal("GetString('greeting') 期望 ok=true")
		}
		if s != "你好" {
			t.Errorf("GetString('greeting') = '%s', 期望 '你好'", s)
		}
	})

	t.Run("GetString 类型不匹配返回 false", func(t *testing.T) {
		cache.Set("number", 42, 0)
		_, ok := cache.GetString("number")
		if ok {
			t.Error("GetString('number') 期望 ok=false（类型不匹配）")
		}
	})

	t.Run("GetInt 类型不匹配返回 false", func(t *testing.T) {
		cache.Set("text", "hello", 0)
		_, ok := cache.GetInt("text")
		if ok {
			t.Error("GetInt('text') 期望 ok=false（类型不匹配）")
		}
	})
}

// TestCacheTTL 测试 TTL 过期
func TestCacheTTL(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	t.Run("永不过期（TTL=0）", func(t *testing.T) {
		cache.Set("forever", "永不过期", 0)
		time.Sleep(10 * time.Millisecond)
		_, ok := cache.Get("forever")
		if !ok {
			t.Error("TTL=0 的 key 不应过期")
		}
	})

	t.Run("短 TTL 后过期", func(t *testing.T) {
		cache.Set("temp", "临时数据", 50*time.Millisecond)
		// 在过期前可以获取
		_, ok := cache.Get("temp")
		if !ok {
			t.Fatal("TTL 过期前 Get() 应成功")
		}
		// 等待 TTL 过期
		time.Sleep(70 * time.Millisecond)
		_, ok = cache.Get("temp")
		if ok {
			t.Error("TTL 过期后 Get() 应返回 false")
		}
	})

	t.Run("Item.IsExpired 方法", func(t *testing.T) {
		now := time.Now()
		tests := []struct {
			name     string
			item     Item
			expired  bool
		}{
			{"零值 ExpiresAt 不过期", Item{ExpiresAt: time.Time{}}, false},
			{"未来时间不过期", Item{ExpiresAt: now.Add(1 * time.Hour)}, false},
			{"过去时间已过期", Item{ExpiresAt: now.Add(-1 * time.Second)}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.item.IsExpired(); got != tt.expired {
					t.Errorf("IsExpired() = %v, 期望 %v", got, tt.expired)
				}
			})
		}
	})
}

// TestCacheDelete 测试删除操作
func TestCacheDelete(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	t.Run("删除存在的 key", func(t *testing.T) {
		cache.Delete("key1")
		_, ok := cache.Get("key1")
		if ok {
			t.Error("删除后 Get('key1') 应返回 false")
		}
		// key2 应不受影响
		_, ok = cache.Get("key2")
		if !ok {
			t.Error("key2 应仍然存在")
		}
	})

	t.Run("删除不存在的 key 不 panic", func(t *testing.T) {
		cache.Delete("nonexist") // 不应 panic
	})
}

// TestCacheClear 测试清空操作
func TestCacheClear(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	cache.Set("a", 1, 0)
	cache.Set("b", 2, 0)
	cache.Set("c", 3, 0)

	if cache.Len() != 3 {
		t.Fatalf("Len() = %d, 期望 3", cache.Len())
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Clear() 后 Len() = %d, 期望 0", cache.Len())
	}

	// 清空后统计也应重置
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("Clear() 后统计应重置，Hits=%d, Misses=%d", stats.Hits, stats.Misses)
	}
}

// TestCacheLen 测试 Len()
func TestCacheLen(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	t.Run("空缓存 Len=0", func(t *testing.T) {
		if cache.Len() != 0 {
			t.Errorf("空缓存 Len() = %d, 期望 0", cache.Len())
		}
	})

	t.Run("添加后 Len 正确", func(t *testing.T) {
		cache.Set("x", 1, 0)
		cache.Set("y", 2, 0)
		if cache.Len() != 2 {
			t.Errorf("Len() = %d, 期望 2", cache.Len())
		}
	})
}

// TestCacheStats 测试统计信息
func TestCacheStats(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	t.Run("初始统计为 0", func(t *testing.T) {
		stats := cache.Stats()
		if stats.Hits != 0 || stats.Misses != 0 {
			t.Errorf("初始统计期望 0，得到 Hits=%d, Misses=%d", stats.Hits, stats.Misses)
		}
	})

	t.Run("命中增加 Hits", func(t *testing.T) {
		cache.Set("stats_test", "val", 0)
		cache.Get("stats_test") // 命中
		stats := cache.Stats()
		if stats.Hits != 1 {
			t.Errorf("Hits 期望 1，得到 %d", stats.Hits)
		}
	})

	t.Run("未命中增加 Misses", func(t *testing.T) {
		cache.Get("no_such_key") // 未命中
		stats := cache.Stats()
		if stats.Misses != 1 {
			t.Errorf("Misses 期望 1，得到 %d", stats.Misses)
		}
	})
}

// TestCacheKeys 测试 Keys()
func TestCacheKeys(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	cache.Set("k1", "v1", 0)
	cache.Set("k2", "v2", 0)
	cache.Set("k3", "v3", 0)

	keys := cache.Keys()
	if len(keys) != 3 {
		t.Fatalf("Keys() 长度 = %d, 期望 3", len(keys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, k := range []string{"k1", "k2", "k3"} {
		if !keySet[k] {
			t.Errorf("Keys() 中缺少 %s", k)
		}
	}
}

// TestCacheConcurrency 测试并发安全
func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	var wg sync.WaitGroup
	n := 50

	// 并发写
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx%26))
			cache.Set(key, idx, 0)
		}(i)
	}

	// 并发读
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := string(rune('A' + idx%26))
			cache.Get(key)
		}(i)
	}

	wg.Wait()
	// 测试不 panic 即通过
}

// TestCacheCleanupLoop 测试定期清理协程
func TestCacheCleanupLoop(t *testing.T) {
	cache := NewCache(50 * time.Millisecond) // 每 50ms 清理一次
	defer cache.Stop()

	cache.Set("short", "很快过期", 30*time.Millisecond)
	cache.Set("long", "长期有效", 1*time.Hour)

	// 等待短 TTL 过期并被清理
	time.Sleep(100 * time.Millisecond)

	_, ok := cache.Get("short")
	if ok {
		t.Error("短 TTL key 应在清理后不可见")
	}

	_, ok = cache.Get("long")
	if !ok {
		t.Error("长 TTL key 应在清理后仍然存在")
	}
}

// TestCacheOverwrite 测试覆盖已存在的 key
func TestCacheOverwrite(t *testing.T) {
	cache := NewCache(0)
	defer cache.Stop()

	cache.Set("key", "旧值", 0)
	cache.Set("key", "新值", 0)

	v, ok := cache.Get("key")
	if !ok {
		t.Fatal("Get('key') 应返回 true")
	}
	if v != "新值" {
		t.Errorf("覆盖后值 = %v, 期望 '新值'", v)
	}
}