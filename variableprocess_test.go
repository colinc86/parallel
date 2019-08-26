package parallel

import (
	"math"
	"testing"
	"time"
)

// MARK: Tests

func TestVariableProcessCompleteness(t *testing.T) {
	v := make([]float64, 10000000)
	c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
	p := NewVariableProcess(100*time.Millisecond, 1, 20, c, false)
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

func TestStopVariableProcess(t *testing.T) {
	v := make([]float64, 1000000)
	c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
	p := NewVariableProcess(100*time.Millisecond, 1, 20, c, false)
	p.Execute(len(v), func(i int) {
		if i == len(v)/2 {
			p.Stop()
		}

		v[i] = float64(i + 1)
	})

	for i, value := range v {
		if i <= len(v)/2 {
			if float64(i+1) != value {
				t.Errorf("Value, %f, should be equal to %f.", value, float64(i+1))
				break
			}
		}
	}
}

// MARK: Benchmarks

func BenchmarkVariableProcess(b *testing.B) {
	v := make([]float64, 1000000)
	c := NewControllerConfiguration(2.0, 0.0, 1.0, 1.0, 1.0)
	p := NewVariableProcess(100*time.Millisecond, 1, 20, c, false)

	for n := 0; n < b.N; n++ {
		p.Execute(len(v), func(i int) {
			v[i] = math.Sqrt(float64(i))
		})
	}
}
