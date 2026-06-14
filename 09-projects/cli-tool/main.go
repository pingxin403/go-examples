package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Task 表示一个待办事项
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// Tasks 管理待办列表
type Tasks struct {
	Items  []Task `json:"items"`
	NextID int    `json:"next_id"`
}

const dataFile = "tasks.json"

// loadTasks 从 JSON 文件加载任务
func loadTasks() (*Tasks, error) {
	t := &Tasks{Items: []Task{}, NextID: 1}
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return t, nil // 文件不存在返回空列表
		}
		return nil, err
	}
	if err := json.Unmarshal(data, t); err != nil {
		return nil, fmt.Errorf("解析任务文件失败: %w", err)
	}
	return t, nil
}

// saveTasks 将任务保存到 JSON 文件
func saveTasks(t *Tasks) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}
	return os.WriteFile(dataFile, data, 0644)
}

// addTask 添加新任务
func addTask(title string) {
	t, err := loadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	task := Task{
		ID:        t.NextID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	t.Items = append(t.Items, task)
	t.NextID++
	if err := saveTasks(t); err != nil {
		fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 已添加任务 #%d: %s\n", task.ID, task.Title)
}

// listTasks 列出所有任务
func listTasks(all bool) {
	t, err := loadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	if len(t.Items) == 0 {
		fmt.Println("📭 暂无任务")
		return
	}
	for _, task := range t.Items {
		if !all && task.Completed {
			continue // 默认只显示未完成
		}
		status := "⬜"
		if task.Completed {
			status = "✅"
		}
		fmt.Printf("%s #%d %s\n", status, task.ID, task.Title)
	}
}

// completeTask 标记任务为已完成
func completeTask(id int) {
	t, err := loadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	for i := range t.Items {
		if t.Items[i].ID == id {
			t.Items[i].Completed = true
			if err := saveTasks(t); err != nil {
				fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ 已完成任务 #%d: %s\n", id, t.Items[i].Title)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "错误: 未找到任务 #%d\n", id)
	os.Exit(1)
}

// deleteTask 删除任务
func deleteTask(id int) {
	t, err := loadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	for i := range t.Items {
		if t.Items[i].ID == id {
			title := t.Items[i].Title
			t.Items = append(t.Items[:i], t.Items[i+1:]...)
			if err := saveTasks(t); err != nil {
				fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("🗑️ 已删除任务 #%d: %s\n", id, title)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "错误: 未找到任务 #%d\n", id)
	os.Exit(1)
}

func main() {
	// 使用 flag 包实现子命令模式
	// 第一个参数是子命令名称
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <command> [options]")
		fmt.Println("命令:")
		fmt.Println("  add <title>      添加新任务")
		fmt.Println("  list [-all]      列出任务")
		fmt.Println("  complete <id>    完成任务")
		fmt.Println("  delete <id>      删除任务")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "add":
		// add 子命令: go run main.go add "买牛奶"
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		addCmd.Parse(os.Args[2:])
		if addCmd.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "错误: 请提供任务标题")
			fmt.Println("用法: go run main.go add <title>")
			os.Exit(1)
		}
		title := strings.Join(addCmd.Args(), " ")
		addTask(title)

	case "list":
		// list 子命令: go run main.go list [-all]
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		all := listCmd.Bool("all", false, "显示所有任务（包括已完成）")
		listCmd.Parse(os.Args[2:])
		listTasks(*all)

	case "complete":
		// complete 子命令: go run main.go complete 1
		completeCmd := flag.NewFlagSet("complete", flag.ExitOnError)
		completeCmd.Parse(os.Args[2:])
		if completeCmd.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "错误: 请提供任务 ID")
			fmt.Println("用法: go run main.go complete <id>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(completeCmd.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID: %s\n", completeCmd.Arg(0))
			os.Exit(1)
		}
		completeTask(id)

	case "delete":
		// delete 子命令: go run main.go delete 1
		deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
		deleteCmd.Parse(os.Args[2:])
		if deleteCmd.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "错误: 请提供任务 ID")
			fmt.Println("用法: go run main.go delete <id>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(deleteCmd.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID: %s\n", deleteCmd.Arg(0))
			os.Exit(1)
		}
		deleteTask(id)

	default:
		fmt.Fprintf(os.Stderr, "错误: 未知命令 %q\n", cmd)
		os.Exit(1)
	}
}