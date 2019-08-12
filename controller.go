package parallel

import (
	"math"
	"runtime"
)

// controller types represent a PID controller to control a process.
type controller struct {
	reporter      *reporter
	previousError float64
	totalError    float64
	cpuCount      int
}

// newController creates and resturns a new controller.
func newController() *controller {
	return &controller{
		reporter: newReporter(),
		cpuCount: runtime.NumCPU(),
	}
}

// next calculates the next value from the control loop.
func (c *controller) next() int {
	usage, t := c.reporter.usage()

	p := 1.0 - (usage / float64(c.cpuCount))
	p = (0.01 * p) + (0.99 * c.previousError)

	i := c.totalError + p

	d := (p - c.previousError) / t

	u := 20.0*p + 0.01*i - 3.0*d
	n := int(math.Round(u))

	c.previousError = p
	c.totalError += p

	return n
}

// reset resets the controller's variables.
func (c *controller) reset() {
	c.reporter.reset()
	c.previousError = 0.0
	c.totalError = 0.0
	c.cpuCount = runtime.NumCPU()
}
