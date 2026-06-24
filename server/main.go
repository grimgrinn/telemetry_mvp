package main

import (
	"context"
	"log"
	"net"

	pb "telemetry_mvp/api"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMetricsServiceServer
}

func (s *server) SendMetrics(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {
	metric := req.GetMetric()

	log.Printf("[Получено] Устройство: %s, Значение: %f, Тип: %s", metric.DeviceId, metric.Value, metric.Type)

	return &pb.MetricResponse{Success: true, Message: "Metric received"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMetricsServiceServer(grpcServer, &server{})

	log.Println("gRPC сервер запущен на порту :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при работе сервера: %v", err)
	}
}
