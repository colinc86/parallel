package parallel

// Operation types represent a single operation in a parallel process.
// Responders should perform the i-th operation.
type Operation func(i int)

type finishedProcess struct{}

// Process types execute a specified number of operations on a given number of
// goroutines.
type Process interface {
	Execute(iterations int, operation Operation)
	NumRoutines() int
}
