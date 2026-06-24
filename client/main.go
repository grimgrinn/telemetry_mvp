package main

import (
	"context"
	"log"
	"time"

	pb "telemetry_mvp/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)

	req := &pb.MetricRequest{
		Metric: &pb.Metric{
			DeviceId:  "sensor-001",
			Timestamp: time.Now().Unix(),
			Value:     10.99,
			Type:      "vibration",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SendMetrics(ctx, req)
	if err != nil {
		log.Fatalf("Ошибка вызова: %v", err)
	}

	log.Printf("Ответ сервера: %v", resp.Message)

}
