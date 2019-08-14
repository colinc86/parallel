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
	kp            float64
	ki            float64
	kd            float64
}

// newController creates and resturns a new controller.
func newController(kp float64, ki float64, kd float64) *controller {
	return &controller{
		reporter: newReporter(),
		cpuCount: runtime.NumCPU(),
		kp:       kp,
		ki:       ki,
		kd:       kd,
	}
}

// next calculates the next value from the control loop.
func (c *controller) next() int {
	usage := c.reporter.usage()

	p := 1.0 - (usage / float64(c.cpuCount))
	p = (0.1 * p) + (0.9 * c.previousError)

	i := c.totalError + p

	d := p - c.previousError

	u := c.kp*p + c.ki*i + c.kd*d
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
