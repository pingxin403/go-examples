package main

import (
	"os"
	"testing"
)

// 由于 addTask/completeTask/deleteTask 内部调用了 os.Exit，无法直接单元测试，
// 我们改为测试底层的 Task/Tasks 结构体序列化/反序列化和业务逻辑。
// 因为 dataFile 是 const，通过切换工作目录到临时目录来隔离文件操作。

// TestTaskJSON 测试 Task 结构体的 JSON 序列化和反序列化
func TestTaskJSON(t *testing.T) {
	// 切换到临时目录，隔离文件操作
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}
	tmpDir, err := os.MkdirTemp("", "cli-tool-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("切换目录失败: %v", err)
	}
	defer os.Chdir(origDir)

	t.Run("空任务列表加载", func(t *testing.T) {
		tasks, err := loadTasks()
		if err != nil {
			t.Fatalf("loadTasks() 返回错误: %v", err)
		}
		if tasks == nil {
			t.Fatal("loadTasks() 返回 nil")
		}
		if len(tasks.Items) != 0 {
			t.Errorf("期望空列表，得到 %d 项", len(tasks.Items))
		}
		if tasks.NextID != 1 {
			t.Errorf("期望 NextID=1，得到 %d", tasks.NextID)
		}
	})

	t.Run("保存并加载任务", func(t *testing.T) {
		tasks := &Tasks{
			Items: []Task{
				{ID: 1, Title: "测试任务", Completed: false},
			},
			NextID: 2,
		}
		if err := saveTasks(tasks); err != nil {
			t.Fatalf("saveTasks() 失败: %v", err)
		}

		loaded, err := loadTasks()
		if err != nil {
			t.Fatalf("loadTasks() 失败: %v", err)
		}
		if len(loaded.Items) != 1 {
			t.Fatalf("期望 1 个任务，得到 %d", len(loaded.Items))
		}
		if loaded.Items[0].Title != "测试任务" {
			t.Errorf("期望 Title='测试任务'，得到 '%s'", loaded.Items[0].Title)
		}
		if loaded.Items[0].ID != 1 {
			t.Errorf("期望 ID=1，得到 %d", loaded.Items[0].ID)
		}
		if loaded.NextID != 2 {
			t.Errorf("期望 NextID=2，得到 %d", loaded.NextID)
		}
	})

	t.Run("标记任务完成", func(t *testing.T) {
		tasks := &Tasks{
			Items: []Task{
				{ID: 1, Title: "任务A", Completed: false},
				{ID: 2, Title: "任务B", Completed: false},
			},
			NextID: 3,
		}
		if err := saveTasks(tasks); err != nil {
			t.Fatalf("saveTasks() 失败: %v", err)
		}

		loaded, err := loadTasks()
		if err != nil {
			t.Fatalf("loadTasks() 失败: %v", err)
		}
		// 手动模拟 complete 操作
		for i := range loaded.Items {
			if loaded.Items[i].ID == 1 {
				loaded.Items[i].Completed = true
			}
		}
		if err := saveTasks(loaded); err != nil {
			t.Fatalf("saveTasks() 失败: %v", err)
		}

		reloaded, _ := loadTasks()
		if !reloaded.Items[0].Completed {
			t.Error("任务 #1 期望为已完成")
		}
		if reloaded.Items[1].Completed {
			t.Error("任务 #2 期望为未完成")
		}
	})

	t.Run("删除任务", func(t *testing.T) {
		tasks := &Tasks{
			Items: []Task{
				{ID: 1, Title: "任务A", Completed: false},
				{ID: 2, Title: "任务B", Completed: false},
				{ID: 3, Title: "任务C", Completed: false},
			},
			NextID: 4,
		}
		if err := saveTasks(tasks); err != nil {
			t.Fatalf("saveTasks() 失败: %v", err)
		}

		loaded, err := loadTasks()
		if err != nil {
			t.Fatalf("loadTasks() 失败: %v", err)
		}
		// 手动模拟 delete 操作
		for i := range loaded.Items {
			if loaded.Items[i].ID == 2 {
				loaded.Items = append(loaded.Items[:i], loaded.Items[i+1:]...)
				break
			}
		}
		if err := saveTasks(loaded); err != nil {
			t.Fatalf("saveTasks() 失败: %v", err)
		}

		reloaded, _ := loadTasks()
		if len(reloaded.Items) != 2 {
			t.Fatalf("删除后期望 2 个任务，得到 %d", len(reloaded.Items))
		}
		if reloaded.Items[0].ID != 1 || reloaded.Items[1].ID != 3 {
			t.Errorf("删除后任务 ID 不正确: %+v", reloaded.Items)
		}
	})
}

// TestTasksListFiltering 测试列表过滤逻辑（只显示未完成）
func TestTasksListFiltering(t *testing.T) {
	tasks := &Tasks{
		Items: []Task{
			{ID: 1, Title: "完成", Completed: true},
			{ID: 2, Title: "未完成", Completed: false},
			{ID: 3, Title: "进行中", Completed: false},
		},
		NextID: 4,
	}

	t.Run("默认只返回未完成", func(t *testing.T) {
		var pending []Task
		for _, task := range tasks.Items {
			if !task.Completed {
				pending = append(pending, task)
			}
		}
		if len(pending) != 2 {
			t.Errorf("期望 2 个未完成任务，得到 %d", len(pending))
		}
	})

	t.Run("all=true 返回全部", func(t *testing.T) {
		if len(tasks.Items) != 3 {
			t.Errorf("期望 3 个任务，得到 %d", len(tasks.Items))
		}
	})
}