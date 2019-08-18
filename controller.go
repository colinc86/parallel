package parallel

import (
	"runtime"
)

// controller types represent a PID controller to control a process.
type controller struct {
	previousError  float64
	totalError     float64
	previousOutput float64
	cpuCount       int
	configuration  *ControllerConfiguration
}

// newController creates and resturns a new controller.
func newController(configuration *ControllerConfiguration) *controller {
	return &controller{
		cpuCount:      runtime.NumCPU(),
		configuration: configuration,
	}
}

// next calculates the next controller output signal from input.
func (c *controller) next(input float64) (float64, float64) {
	e := 1.0 - (input / float64(c.cpuCount))
	e = c.configuration.ErrorResponse * e + (1.0 - c.configuration.ErrorResponse) * c.previousError

	i := c.totalError + e

	d := e - c.previousError

	u := c.configuration.Kp*e + c.configuration.Ki*i + c.configuration.Kd*d
	u = c.configuration.OutputResponse * u + (c.configuration.OutputResponse - 1) * c.previousOutput

	c.previousError = e
	c.totalError = i
	c.previousOutput = u

	return u, e
}

// reset resets the controller's variables.
func (c *controller) reset() {
	c.previousError = 0.0
	c.totalError = 0.0
	c.cpuCount = runtime.NumCPU()
}
