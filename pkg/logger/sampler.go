package logger

import (
	"context"
	"sync/atomic"
	"time"
)

type Sampler struct {
	rate      int64
	counter   atomic.Int64
	interval  time.Duration
	lastReset atomic.Int64
}

func NewSampler(rate int, interval time.Duration) *Sampler {
	s := &Sampler{
		rate:     int64(rate),
		interval: interval,
	}

	s.lastReset.Store(time.Now().UnixNano())

	return s
}

func (s *Sampler) ShouldLog() bool {
	now := time.Now().UnixNano()
	lastReset := s.lastReset.Load()

	if now-lastReset > s.interval.Nanoseconds() {
		s.counter.Store(0)
		s.lastReset.Store(now)
	}

	count := s.counter.Add(1)

	return count <= s.rate
}

type SampledLogger struct {
	logger  Interface
	sampler *Sampler
}

func NewSampledLogger(l Interface, rate int, interval time.Duration) *SampledLogger {
	return &SampledLogger{
		logger:  l,
		sampler: NewSampler(rate, interval),
	}
}

func (sl *SampledLogger) Debug(message interface{}, args ...interface{}) {
	if sl.sampler.ShouldLog() {
		sl.logger.Debug(message, args...)
	}
}

func (sl *SampledLogger) Info(message string, args ...interface{}) {
	if sl.sampler.ShouldLog() {
		sl.logger.Info(message, args...)
	}
}

func (sl *SampledLogger) Warn(message string, args ...interface{}) {
	sl.logger.Warn(message, args...)
}

func (sl *SampledLogger) Error(message interface{}, args ...interface{}) {
	sl.logger.Error(message, args...)
}

func (sl *SampledLogger) Fatal(message interface{}, args ...interface{}) {
	sl.logger.Fatal(message, args...)
}

func (sl *SampledLogger) WithField(key string, value interface{}) Interface {
	return &SampledLogger{
		logger:  sl.logger.WithField(key, value),
		sampler: sl.sampler,
	}
}

func (sl *SampledLogger) WithFields(fields map[string]interface{}) Interface {
	return &SampledLogger{
		logger:  sl.logger.WithFields(fields),
		sampler: sl.sampler,
	}
}

func (sl *SampledLogger) WithRequestID(requestID string) Interface {
	return &SampledLogger{
		logger:  sl.logger.WithRequestID(requestID),
		sampler: sl.sampler,
	}
}

func (sl *SampledLogger) WithContext(ctx context.Context) Interface {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		return sl.WithRequestID(requestID)
	}

	return sl
}
