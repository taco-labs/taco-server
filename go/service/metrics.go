package service

import (
	"time"

	"github.com/smira/go-statsd"
	"github.com/taco-labs/taco/go/utils/slices"
	"go.uber.org/zap"
)

// TODO (taekyeom) expand to integer tags..
type Tag struct {
	Key   string
	Value string
}

type MetricService interface {
	// Timing tracks a duration event
	Timing(stat string, duration time.Duration, tags ...Tag)
}

// Do nothing
type mockMetricService struct{}

func (m mockMetricService) Timing(stat string, duration time.Duration, tags ...Tag) {}

func NewMockMetricService() *mockMetricService {
	return &mockMetricService{}
}

type loggerMetricService struct {
	logger *zap.Logger
}

func (m loggerMetricService) Timing(stat string, duration time.Duration, tags ...Tag) {
	tagMaps := make(map[string]string, len(tags))
	for _, tag := range tags {
		tagMaps[tag.Key] = tag.Value
	}

	m.logger.Info("metric",
		zap.String("kind", "duration"),
		zap.String("stat", stat),
		zap.Duration("duration", duration),
		zap.Any("tag", tagMaps))
}

func NewLoggerMetricService(logger *zap.Logger) *loggerMetricService {
	return &loggerMetricService{logger}
}

type StatsDMetricService struct {
	client *statsd.Client
}

func (s StatsDMetricService) Timing(stat string, duration time.Duration, tags ...Tag) {
	statsDTags := slices.Map(tags, func(i Tag) statsd.Tag {
		return statsd.StringTag(i.Key, i.Value)
	})
	s.client.Timing(stat, duration.Milliseconds(), statsDTags...)
}

func NewStatsDMetricService(client *statsd.Client) *StatsDMetricService {
	return &StatsDMetricService{
		client: client,
	}
}
