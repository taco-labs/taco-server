package service

import (
	"time"

	"github.com/smira/go-statsd"
	"go.uber.org/zap"
)

type MetricService interface {
	// Timing tracks a duration event
	Timing(stat string, duration time.Duration, tagKvs ...string)
	Count(stat string, count int64, tagKvs ...string)
}

// Do nothing
type mockMetricService struct{}

func (m mockMetricService) Timing(stat string, duration time.Duration, tagKvs ...string) {}

func (m mockMetricService) Count(stat string, count int64, tagKvs ...string) {}

func NewMockMetricService() *mockMetricService {
	return &mockMetricService{}
}

type loggerMetricService struct {
	logger *zap.Logger
}

func (m loggerMetricService) Timing(stat string, duration time.Duration, tagKvs ...string) {
	if len(tagKvs)%2 == 1 {
		// TODO warning log
	}

	tagMaps := make(map[string]string, len(tagKvs)/2)
	for i := 0; i < len(tagKvs)/2; i++ {
		tagMaps[tagKvs[2*i]] = tagKvs[2*i+1]
	}

	m.logger.Info("metric",
		zap.String("kind", "duration"),
		zap.String("stat", stat),
		zap.Duration("duration", duration),
		zap.Any("tag", tagMaps))
}

func (m loggerMetricService) Count(stat string, count int64, tagKvs ...string) {
	if len(tagKvs)%2 == 1 {
		// TODO warning log
	}

	tagMaps := make(map[string]string, len(tagKvs)/2)
	for i := 0; i < len(tagKvs)/2; i++ {
		tagMaps[tagKvs[2*i]] = tagKvs[2*i+1]
	}

	m.logger.Info("metric",
		zap.String("kind", "count"),
		zap.String("stat", stat),
		zap.Int64("count", count),
		zap.Any("tag", tagMaps))

}

func NewLoggerMetricService(logger *zap.Logger) *loggerMetricService {
	return &loggerMetricService{logger}
}

type StatsDMetricService struct {
	client *statsd.Client
}

func (s StatsDMetricService) Timing(stat string, duration time.Duration, tagKvs ...string) {
	if len(tagKvs)%2 == 1 {
		// TODO warning log
	}

	statsDTags := make([]statsd.Tag, len(tagKvs)/2, len(tagKvs)/2)
	for i := 0; i < len(tagKvs)/2; i++ {
		statsDTags[i] = statsd.StringTag(tagKvs[2*i], tagKvs[2*i+1])
	}
	s.client.Timing(stat, duration.Milliseconds(), statsDTags...)
}

func (s StatsDMetricService) Count(stat string, count int64, tagKvs ...string) {
	if len(tagKvs)%2 == 1 {
		// TODO warning log
	}

	statsDTags := make([]statsd.Tag, len(tagKvs)/2, len(tagKvs)/2)
	for i := 0; i < len(tagKvs)/2; i++ {
		statsDTags[i] = statsd.StringTag(tagKvs[2*i], tagKvs[2*i+1])
	}
	s.client.Incr(stat, count, statsDTags...)
}

func NewStatsDMetricService(client *statsd.Client) *StatsDMetricService {
	return &StatsDMetricService{
		client: client,
	}
}
