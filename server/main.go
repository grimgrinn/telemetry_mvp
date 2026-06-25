package main

import (
	"context"
	"log"
	"net"

	pb "telemetry_mvp/api"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"

	"telemetry_mvp/config"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type server struct {
	pb.UnimplementedMetricsServiceServer
	producer sarama.SyncProducer
}

func (s *server) SendMetrics(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {
	metric := req.GetMetric()

	log.Printf("[Получено] Устройство: %s, Значение: %f, Тип: %s", metric.DeviceId, metric.Value, metric.Type)

	data, err := proto.Marshal(metric)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return &pb.MetricResponse{Success: false, Message: "marshal error"}, nil
	}

	msg := &sarama.ProducerMessage{
		Topic: "telemetry",
		Key:   sarama.StringEncoder(metric.DeviceId),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = s.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Kafka error: %v", err)
		return &pb.MetricResponse{Success: false, Message: "kafka error"}, nil
	}

	log.Printf("Отправлено в Kafka: %s", metric.DeviceId)
	return &pb.MetricResponse{Success: true, Message: "Metric received"}, nil
}

func main() {
	cfg := config.Load()
	log.Printf("Config loaded: port=%s, log_level=%s", cfg.ServerPort, cfg.LogLevel)

	// kafka
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5
	kafkaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{cfg.KafkaBroker}, kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()
	log.Printf("Kafka producer connected to %s", cfg.KafkaBroker)

	// grpc
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServiceServer(grpcServer, &server{producer: producer})

	log.Println("gRPC сервер запущен на порту :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при работе сервера: %v", err)
	}
}
