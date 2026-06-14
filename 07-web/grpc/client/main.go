// main.go — gRPC 客户端示例
//
// 本文件实现 gRPC HelloService 的客户端，演示：
//   - 连接 gRPC 服务器
//   - 调用简单 RPC (SayHello)
//   - 调用服务端流式 RPC (SayHelloStream)
//   - 带超时的 Context
//   - 连接配置 (TLS、拦截器等)
//
// 安装依赖:
//   go get google.golang.org/grpc google.golang.org/protobuf
//
// 运行前需要先启动服务端:
//   go run 07-web/grpc/server/main.go
//
// 然后运行客户端:
//   go run 07-web/grpc/client/main.go

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/pingxin403/go-examples/07-web/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// =========================================================
// 1. 简单 RPC 调用
// =========================================================

// callSayHello 调用简单 RPC SayHello
func callSayHello(client pb.HelloServiceClient, name string) {
	log.Printf(">>> 调用 SayHello(name=%s)", name)

	// 设置 5 秒超时的 Context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发起 RPC 调用
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("SayHello 调用失败: %v", err)
	}

	log.Printf("<<< 响应: %s\n", resp.GetMessage())
}

// =========================================================
// 2. 服务端流式 RPC 调用
// =========================================================

// callSayHelloStream 调用服务端流式 RPC SayHelloStream
func callSayHelloStream(client pb.HelloServiceClient, name string) {
	log.Printf(">>> 调用 SayHelloStream(name=%s)", name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 发起流式 RPC 调用，返回一个 stream
	stream, err := client.SayHelloStream(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("SayHelloStream 启动失败: %v", err)
	}

	log.Println("<<< 接收流式响应:")

	// 循环接收服务端推送的消息
	index := 1
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			// 服务端关闭流
			break
		}
		if err != nil {
			log.Fatalf("SayHelloStream 接收失败: %v", err)
		}
		log.Printf("  [%d] %s", index, resp.GetMessage())
		index++
	}

	log.Printf("流式 RPC 完成，共收到 %d 条消息\n", index-1)
}

// =========================================================
// 3. 批量调用演示
// =========================================================

// batchCall 批量调用 SayHello
func batchCall(client pb.HelloServiceClient, names []string) {
	log.Println(">>> 批量调用 SayHello")

	for _, name := range names {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("SayHello('%s') 失败: %v", name, err)
		} else {
			log.Printf("  %s → %s", name, resp.GetMessage())
		}
		cancel()
	}

	log.Println("批量调用完成")
}

// =========================================================
// 4. 主函数
// =========================================================

func main() {
	// =========================================================
	// 连接 gRPC 服务器
	// =========================================================
	serverAddr := "localhost:50051"

	log.Printf("正在连接 gRPC 服务器 %s ...", serverAddr)

	// 建立连接（开发环境使用不安全的明文连接）
	// 生产环境应使用 TLS: grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(...))
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),                 // 阻塞直到连接建立
		grpc.WithTimeout(5*time.Second),  // 连接超时
	)
	if err != nil {
		log.Fatalf("连接 gRPC 服务器失败: %v", err)
	}
	defer conn.Close()

	log.Printf("成功连接到 %s\n", serverAddr)

	// 创建客户端 Stub
	client := pb.NewHelloServiceClient(conn)

	// =========================================================
	// 调用示例
	// =========================================================
	fmt.Println("═══════════ gRPC 客户端演示 ═══════════")
	fmt.Println()

	// 1. 简单 RPC
	callSayHello(client, "张三")

	// 2. 流式 RPC
	callSayHelloStream(client, "李四")

	// 3. 批量调用
	batchCall(client, []string{"Alice", "Bob", "Charlie"})

	fmt.Println("\n═══════════ 所有调用完成 ═══════════")
}