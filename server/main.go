package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	pb "telemetry_mvp/api"

	_ "github.com/lib/pq"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMetricsServiceServer
	db *sql.DB
}

func (s *server) SendMetrics(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {
	metric := req.GetMetric()

	log.Printf("[Получено] Устройство: %s, Значение: %f, Тип: %s", metric.DeviceId, metric.Value, metric.Type)

	// database
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO metrics (device_id, timestamp, value, type) VALUES ($1, $2, $3, $4)",
		metric.DeviceId, metric.Timestamp, metric.Value, metric.Type)
	if err != nil {
		log.Printf("DB error: %v", err)
		return &pb.MetricResponse{Success: false, Message: "DB error"}, nil
	}

	return &pb.MetricResponse{Success: true, Message: "Metric received"}, nil
}

func main() {
	// db
	connStr := "postgres://user:pass@localhost:5432/telemetry?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id SERIAL PRIMARY KEY,
			device_id TEXT,
			timestamp BIGINT,
			value FLOAT,
			type  TEXt
		);
	`)
	if err != nil {
		log.Fatalf("Failed to creat table: %v", err)
	}
	log.Println("DB connected and table ready")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMetricsServiceServer(grpcServer, &server{db: db})

	log.Println("gRPC сервер запущен на порту :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при работе сервера: %v", err)
	}
}
