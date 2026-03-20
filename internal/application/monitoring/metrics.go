package monitoring

import (
	"fmt"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const MetricKeyFrpsDelay = "frps_delay"

// MetricsService handles writing latency data to fh_metrics_targets and fh_metrics_raw.
type MetricsService struct {
	app core.App

	// targetCache maps serverID -> fh_metrics_targets record ID
	targetCache     map[string]string
	targetCacheLock sync.RWMutex
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(app core.App) *MetricsService {
	return &MetricsService{
		app:         app,
		targetCache: make(map[string]string),
	}
}

// RecordServerLatency records a latency measurement for a server.
// It automatically finds or creates the corresponding fh_metrics_targets record.
// Negative latency (unreachable host) is skipped.
func (s *MetricsService) RecordServerLatency(serverID string, latencyMs int64) error {
	if latencyMs < 0 {
		return nil
	}

	targetID, err := s.getOrCreateServerTarget(serverID)
	if err != nil {
		return fmt.Errorf("get/create metrics target: %w", err)
	}

	return s.recordMetric(targetID, MetricKeyFrpsDelay, float64(latencyMs))
}

// getOrCreateServerTarget returns the fh_metrics_targets record ID for a server,
// creating one if it does not exist. Results are cached in memory.
func (s *MetricsService) getOrCreateServerTarget(serverID string) (string, error) {
	s.targetCacheLock.RLock()
	if id, ok := s.targetCache[serverID]; ok {
		s.targetCacheLock.RUnlock()
		return id, nil
	}
	s.targetCacheLock.RUnlock()

	// Acquire write lock for the entire find-or-create to prevent duplicate records
	// when multiple goroutines encounter a cache miss simultaneously.
	s.targetCacheLock.Lock()
	defer s.targetCacheLock.Unlock()

	// Re-check under write lock (another goroutine may have just created it).
	if id, ok := s.targetCache[serverID]; ok {
		return id, nil
	}

	records, err := s.app.FindRecordsByFilter(
		"fh_metrics_targets",
		"serverId = {:serverId} && proxyId = ''",
		"",
		1,
		0,
		map[string]any{"serverId": serverID},
	)
	if err != nil {
		return "", fmt.Errorf("query metrics target: %w", err)
	}

	var targetID string
	if len(records) > 0 {
		targetID = records[0].Id
	} else {
		collection, err := s.app.FindCollectionByNameOrId("fh_metrics_targets")
		if err != nil {
			return "", fmt.Errorf("find collection fh_metrics_targets: %w", err)
		}
		record := core.NewRecord(collection)
		record.Set("serverId", serverID)
		if err := s.app.Save(record); err != nil {
			return "", fmt.Errorf("save metrics target: %w", err)
		}
		targetID = record.Id
	}

	s.targetCache[serverID] = targetID
	return targetID, nil
}

// GetLatestLatencyBatch returns the most recent latency for every server in one SQL query.
// The returned map key is the server ID.
func (s *MetricsService) GetLatestLatencyBatch() (map[string]*NetworkStatus, error) {
	type row struct {
		ServerID string  `db:"serverId"`
		Val      float64 `db:"val"`
		T        string  `db:"t"`
	}

	var rows []row
	err := s.app.DB().
		NewQuery(`
			SELECT t.serverId, r.val, r.t
			FROM fh_metrics_raw r
			INNER JOIN fh_metrics_targets t ON r.target_id = t.id
			WHERE r.metric_key = {:key}
			  AND r.t = (
			    SELECT MAX(r2.t) FROM fh_metrics_raw r2
			    WHERE r2.target_id = r.target_id AND r2.metric_key = {:key}
			  )
		`).
		Bind(dbx.Params{"key": MetricKeyFrpsDelay}).
		All(&rows)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*NetworkStatus, len(rows))
	for _, row := range rows {
		latency := int64(row.Val)
		t, _ := time.Parse("2006-01-02 15:04:05.000Z", row.T)
		result[row.ServerID] = &NetworkStatus{
			Latency:       latency,
			Reachable:     latency >= 0,
			LastCheckTime: t,
		}
	}
	return result, nil
}

// GetLatestServerLatency returns the most recent latency measurement for a server.
// Returns nil if no data has been recorded yet.
func (s *MetricsService) GetLatestServerLatency(serverID string) (*NetworkStatus, error) {
	records, err := s.app.FindRecordsByFilter(
		"fh_metrics_raw",
		"target_id.serverId = {:serverId} && metric_key = {:key}",
		"-t",
		1,
		0,
		map[string]any{"serverId": serverID, "key": MetricKeyFrpsDelay},
	)
	if err != nil {
		return nil, fmt.Errorf("query latest latency: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	r := records[0]
	latency := int64(r.GetFloat("val"))
	return &NetworkStatus{
		Latency:       latency,
		Reachable:     latency >= 0,
		LastCheckTime: r.GetDateTime("t").Time(),
	}, nil
}

// recordMetric writes a single data point to fh_metrics_raw.
func (s *MetricsService) recordMetric(targetID, metricKey string, value float64) error {
	collection, err := s.app.FindCollectionByNameOrId("fh_metrics_raw")
	if err != nil {
		return fmt.Errorf("find collection fh_metrics_raw: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("target_id", targetID)
	record.Set("metric_key", metricKey)
	record.Set("t", time.Now().UTC().Format("2006-01-02 15:04:05.000Z"))
	record.Set("val", value)

	if err := s.app.Save(record); err != nil {
		return fmt.Errorf("save metric record: %w", err)
	}

	return nil
}
