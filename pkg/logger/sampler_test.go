package logger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSampler(t *testing.T) {
	t.Parallel()

	s := NewSampler(10, time.Second)

	require.NotNil(t, s)
	assert.Equal(t, int64(10), s.rate)
	assert.Equal(t, time.Second, s.interval)
}

func TestSampler_ShouldLog(t *testing.T) {
	t.Parallel()

	t.Run("allows logs up to rate", func(t *testing.T) {
		t.Parallel()

		s := NewSampler(3, time.Minute)

		assert.True(t, s.ShouldLog())
		assert.True(t, s.ShouldLog())
		assert.True(t, s.ShouldLog())
		assert.False(t, s.ShouldLog())
		assert.False(t, s.ShouldLog())
	})

	t.Run("resets after interval", func(t *testing.T) {
		t.Parallel()

		s := NewSampler(1, 10*time.Millisecond)

		assert.True(t, s.ShouldLog())
		assert.False(t, s.ShouldLog())

		time.Sleep(15 * time.Millisecond)

		assert.True(t, s.ShouldLog())
	})
}

func TestNewSampledLogger(t *testing.T) {
	t.Parallel()

	l := New("info")
	sl := NewSampledLogger(l, 100, time.Second)

	require.NotNil(t, sl)
	assert.NotNil(t, sl.logger)
	assert.NotNil(t, sl.sampler)
}

func TestSampledLogger_WithField(t *testing.T) {
	t.Parallel()

	l := New("info")
	sl := NewSampledLogger(l, 100, time.Second)

	result := sl.WithField("key", "value")

	require.NotNil(t, result)

	_, ok := result.(*SampledLogger)
	assert.True(t, ok)
}

func TestSampledLogger_WithFields(t *testing.T) {
	t.Parallel()

	l := New("info")
	sl := NewSampledLogger(l, 100, time.Second)

	result := sl.WithFields(map[string]interface{}{"key": "value"})

	require.NotNil(t, result)

	_, ok := result.(*SampledLogger)
	assert.True(t, ok)
}

func TestSampledLogger_WithRequestID(t *testing.T) {
	t.Parallel()

	l := New("info")
	sl := NewSampledLogger(l, 100, time.Second)

	result := sl.WithRequestID("test-id")

	require.NotNil(t, result)

	_, ok := result.(*SampledLogger)
	assert.True(t, ok)
}

func TestSampledLogger_WithContext(t *testing.T) {
	t.Parallel()

	l := New("info")
	sl := NewSampledLogger(l, 100, time.Second)

	t.Run("with request ID", func(t *testing.T) {
		t.Parallel()

		ctx := ContextWithRequestID(context.Background(), "ctx-id")

		result := sl.WithContext(ctx)

		require.NotNil(t, result)

		_, ok := result.(*SampledLogger)
		assert.True(t, ok)
	})

	t.Run("without request ID", func(t *testing.T) {
		t.Parallel()

		result := sl.WithContext(context.Background())

		assert.Equal(t, sl, result)
	})
}
