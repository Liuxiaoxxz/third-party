package main

import (
	"context"
	"fmt"
	"github.com/Liuxiaoxxz/third-party/grpc/metrics"
	"google.golang.org/grpc"
	"log"
)

func main() {
	// 连接到 gRPC 服务器
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := metrics.NewGrpcClient(conn)
	log.Printf("grpc connection ...")

	res, err := client.Export(context.Background(), &metrics.ExportRequest{
		State: 1,
		Orig: &metrics.ExportMetricsServiceRequest{
			AppName: "hello java!",
		},
	})

	if err != nil {
		log.Fatalf("could not get example: %v", err)
	}

	fmt.Println(res)
}
