package main

import (
	"context"
	"log"
	"net"

	pb "telemetry_mvp/api"

	_ "github.com/lib/pq"

	"telemetry_mvp/config"
	"telemetry_mvp/repository"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMetricsServiceServer
	repo *repository.MetricRepository
}

func (s *server) SendMetrics(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {
	metric := req.GetMetric()

	log.Printf("[Получено] Устройство: %s, Значение: %f, Тип: %s", metric.DeviceId, metric.Value, metric.Type)

	if err := s.repo.Save(ctx, metric); err != nil {
		log.Printf("DB error: %v", err)
		return &pb.MetricResponse{Success: false, Message: "DB error"}, nil
	}

	return &pb.MetricResponse{Success: true, Message: "Metric received"}, nil
}

func main() {
	cfg := config.Load()
	log.Printf("Config loaded: port=%s, log_level=%s", cfg.ServerPort, cfg.LogLevel)

	repo, err := repository.NewMetricRepository(cfg.DBConn)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	log.Println("DB conneceted and talbe ready")

	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMetricsServiceServer(grpcServer, &server{repo: repo})

	log.Println("gRPC сервер запущен на порту :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при работе сервера: %v", err)
	}
}
