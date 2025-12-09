package logger

import (
	"context"
	"sync/atomic"
	"time"
)

// Sampler provides rate-limited sampling for log entries.
// It allows a configurable number of log entries per time interval.
type Sampler struct {
	rate      int64
	counter   atomic.Int64
	interval  time.Duration
	lastReset atomic.Int64
}

// NewSampler creates a new Sampler that allows up to rate log entries per interval.
// Panics if rate or interval is not positive.
func NewSampler(rate int, interval time.Duration) *Sampler {
	if rate <= 0 {
		panic("sampler rate must be positive")
	}

	if interval <= 0 {
		panic("sampler interval must be positive")
	}

	s := &Sampler{
		rate:     int64(rate),
		interval: interval,
	}

	s.lastReset.Store(time.Now().UnixNano())

	return s
}

// ShouldLog returns true if the log entry should be emitted based on sampling rate.
func (s *Sampler) ShouldLog() bool {
	now := time.Now().UnixNano()
	lastReset := s.lastReset.Load()

	if now-lastReset > s.interval.Nanoseconds() {
		if s.lastReset.CompareAndSwap(lastReset, now) {
			s.counter.Store(0)
		}
	}

	count := s.counter.Add(1)

	return count <= s.rate
}

// SampledLogger wraps a logger with rate-limited sampling.
// Only Debug and Info logs are sampled; Warn, Error, and Fatal
// always pass through to ensure important messages are not dropped.
type SampledLogger struct {
	logger  Interface
	sampler *Sampler
}

// NewSampledLogger creates a new SampledLogger that rate-limits Debug and Info logs.
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
