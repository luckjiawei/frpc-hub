package persistence

import (
	"podux/internal/domain/server"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type ServerRepository struct {
	app core.App
}

func NewServerRepository(app core.App) *ServerRepository {
	return &ServerRepository{app: app}
}

func (r *ServerRepository) GetMaxLatency() (int64, error) {
	var maxLatency int64
	err := r.app.DB().
		Select("MAX(val)").
		From("fh_metrics_raw").
		Where(dbx.NewExp("metricKey = 'frps_delay' AND t IN (SELECT MAX(t) FROM fh_metrics_raw WHERE metricKey = 'frps_delay' GROUP BY targetId)")).
		Row(&maxLatency)

	if err != nil {
		r.app.Logger().Error("Failed to get max latency", "error", err)
		return 0, err
	}
	return maxLatency, nil
}

func (r *ServerRepository) FindAllWithAutoConnect() ([]server.ServerRecord, error) {
	records, err := r.app.FindAllRecords("fh_servers", dbx.HashExp{"autoConnection": true})
	if err != nil {
		return nil, err
	}

	result := make([]server.ServerRecord, 0, len(records))
	for _, rec := range records {
		result = append(result, server.ServerRecord{
			ID:         rec.Id,
			ServerName: rec.GetString("serverName"),
		})
	}
	return result, nil
}
