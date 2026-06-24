package repository

import (
	"context"
	"database/sql"

	pb "telemetry_mvp/api"

	_ "github.com/lib/pq"
)

type MetricRepository struct {
	db *sql.DB
}

func NewMetricRepository(connStr string) (*MetricRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &MetricRepository{db: db}, nil
}

func (r *MetricRepository) Save(ctx context.Context, m *pb.Metric) error {
	const query = `INSERT INTO metrics (device_id, timestamp, value, type) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, m.DeviceId, m.Timestamp, m.Value, m.Type)
	return err
}

func (r *MetricRepository) Close() error {
	return r.db.Close()
}
