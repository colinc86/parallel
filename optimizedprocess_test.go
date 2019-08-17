package parallel

import (
	"math"
	"testing"
	"time"
)

// MARK: Tests

func TestOptimizedProcessCompleteness(t *testing.T) {
	v := make([]float64, 1000000)
	c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
	p := NewOptimizedProcess(100*time.Millisecond, 20, c)
	p.Execute(len(v), func(i int) {
		v[i] = float64(i + 1)
	})

	for i, value := range v {
		if float64(i+1) != value {
			t.Errorf("Value, %f, should be equal to %f.", value, float64(i+1))
			break
		}
	}
}

// MARK: Benchmarks

func BenchmarkOptimizedProcess(b *testing.B) {
	v := make([]float64, 1000000)
	c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
	p := NewOptimizedProcess(100*time.Millisecond, 20, c)

	for n := 0; n < b.N; n++ {
		p.Execute(len(v), func(i int) {
			v[i] = math.Sqrt(float64(i))
		})
	}
}
