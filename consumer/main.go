package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	pb "telemetry_mvp/api"
	"telemetry_mvp/config"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBConn)
	if err != nil {
		log.Fatal("DB error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id SERIAL PRIMARY KEY,
			device_id TEXT
			timestamp BIGINT,
			value FLOAT,
			type TEXT
			);
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	log.Println("DB connected and table ready")

	// KAFKA CONSUMER
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer([]string{cfg.KafkaBroker}, kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("telemetry", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	log.Println("Consumer starred. Waitiing for messages...")

	// process messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				var metric pb.Metric
				if err := proto.Unmarshal(msg.Value, &metric); err != nil {
					log.Printf("Failed to unmarshal: %v", err)
					continue
				}

				// write to db
				_, err := db.ExecContext(ctx,
					"INSERT INTO metrics (device_id, timestamp, value, type) VALUES ($1, $2, $3, $4)",
					metric.DeviceId, metric.Timestamp, metric.Value, metric.Type,
				)
				if err != nil {
					log.Printf("DB insert error: %v", err)
				} else {
					log.Printf("Saved to db: %s - %f", metric.DeviceId, metric.Value)
				}
			case err := <-partitionConsumer.Errors():
				log.Printf("Kafka consumer error: %v", err)
			case <-ctx.Done():
				log.Println("Consumer stopping...")
				return
			}
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	cancel()
	log.Println("Consumer stopped")
}
