package parallel

import (
	"math"
	"testing"
)

// MARK: Tests

func TestSyncedProcessCompleteness_01(t *testing.T) {
	v := make([]float64, 1000000)
	p := NewSyncedProcess(1)
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

func TestSyncedProcessCompleteness_02(t *testing.T) {
	v := make([]float64, 1000000)
	p := NewSyncedProcess(2)
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

func BenchmarkSyncedProcess_01(b *testing.B) {
	v := make([]float64, 1000000)
	p := NewSyncedProcess(1)

	for n := 0; n < b.N; n++ {
		p.Execute(len(v), func(i int) {
			v[i] = math.Sqrt(float64(i))
		})
	}
}

func BenchmarkSyncedProcess_02(b *testing.B) {
	v := make([]float64, 1000000)
	p := NewSyncedProcess(2)

	for n := 0; n < b.N; n++ {
		p.Execute(len(v), func(i int) {
			v[i] = math.Sqrt(float64(i))
		})
	}
}
