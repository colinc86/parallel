package parallel

import (
	"math"
	"runtime"
)

// controller types represent a PID controller to control a process.
type controller struct {
	reporter       *reporter
	previousError  float64
	totalError     float64
	previousOutput float64
	cpuCount       int
	configuration  *ControllerConfiguration
}

// newController creates and resturns a new controller.
func newController(configuration *ControllerConfiguration) *controller {
	return &controller{
		reporter:      newReporter(),
		cpuCount:      runtime.NumCPU(),
		configuration: configuration,
	}
}

// next calculates the next value from the control loop.
func (c *controller) next() int {
	usage := c.reporter.usage()

	p := 1.0 - (usage / float64(c.cpuCount))
	p = ((1.0 - c.configuration.ErrorResponse) * p) + (c.configuration.ErrorResponse * c.previousError)

	i := c.totalError + p

	d := p - c.previousError

	u := c.configuration.Kp*p + c.configuration.Ki*i + c.configuration.Kd*d
	u = ((1.0 - c.configuration.OutputResponse) * u) + (c.configuration.OutputResponse * c.previousOutput)

	n := int(math.Round(u))

	c.previousError = p
	c.totalError += p
	c.previousOutput = u

	return n
}

// reset resets the controller's variables.
func (c *controller) reset() {
	c.reporter.reset()
	c.previousError = 0.0
	c.totalError = 0.0
	c.cpuCount = runtime.NumCPU()
}
