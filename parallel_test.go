package parallel

import (
	"math/rand"
	"testing"

	"github.com/colinc86/gdsp"
)

// MARK: Helpers

const vectorLength int = 1000000
const fftLength int = 8

// newVector creates and returns a new vector with length
// defined by fftLength.
func newVector() gdsp.Vector {
	v := make([]float64, vectorLength)
	for i := 0; i < vectorLength; i++ {
		v[i] = rand.Float64()
	}
	return gdsp.MakeVectorFromArray(v)
}

func performFFT(i int, v gdsp.Vector, m []gdsp.VectorComplex) {
	window := v[i : i+fftLength]
	m[i] = gdsp.FFT(window.ToComplex())
}

// // sepectrogram creates a spectrogram from the vector, v, stores the
// // output in the matrix, m, and sends a message to channel c with
// // the number of vectors that were written to m.
func spectrogram(c chan int, v gdsp.Vector, m []gdsp.VectorComplex) {
	for i := 0; i < len(v)-fftLength; i++ {
		performFFT(i, v, m)
	}
	c <- len(v) - fftLength
}

// MARK: Benchmarks - Goroutines

func BenchmarkSum_01(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	c := make(chan int)

	for n := 0; n < b.N; n++ {
		go spectrogram(c, v, s)
		<-c
	}
}

func BenchmarkSum_02(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	c := make(chan int, 2)

	for n := 0; n < b.N; n++ {
		go spectrogram(c, v[:len(s)/2-fftLength], s)
		go spectrogram(c, v[len(s)/2:], s[len(s)/2:])
		<-c
		<-c
	}
}

func BenchmarkSum_03(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	c := make(chan int)

	for n := 0; n < b.N; n++ {
		go spectrogram(c, v[:len(s)/4+fftLength], s)
		go spectrogram(c, v[len(s)/4:len(s)/2], s[len(s)/4:])
		go spectrogram(c, v[len(s)/2:3*len(s)/4], s[len(s)/2:])
		go spectrogram(c, v[3*len(s)/4:], s[3*len(s)/4:])
		<-c
		<-c
		<-c
		<-c
	}
}

// MARK: Benchmarks - Process

func BenchmarkSum_04(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewProcess(1)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

func BenchmarkSum_05(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewProcess(2)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

func BenchmarkSum_06(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewProcess(4)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

func BenchmarkSum_07(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewProcess(10)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

// MARK: Benchmarks - SyncedProcess

func BenchmarkSum_08(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewSyncedProcess(1)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

func BenchmarkSum_09(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewSyncedProcess(2)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

func BenchmarkSum_10(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)
	p := NewSyncedProcess(4)

	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}

// MARK: Benchmarks - OptimizedProcess

func BenchmarkSum_11(b *testing.B) {
	v := newVector()
	s := make([]gdsp.VectorComplex, vectorLength-fftLength)

	p := NewOptimizedProcess(500)
	for n := 0; n < b.N; n++ {
		p.Execute(len(s), func(i int) {
			performFFT(i, v, s)
		})

		<-p.C
	}
}
