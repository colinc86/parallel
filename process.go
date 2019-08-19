// Package parallel provides objects for executing parallel operations on either
// a fixed or variable number of goroutines.
package parallel

// Operation types represent a single operation in a parallel process.
// Responders should perform the i-th operation.
type Operation func(i int)

// Process types execute a specified number of operations on a given number of
// goroutines.
type Process interface {

	// Execute executes a parallel process for the given number of iterations
	// using the provided operation function.
	Execute(iterations int, operation Operation)

	// NumRoutines returns the number of routines that are currently executing in
	// the parallel process.
	NumRoutines() int
}
