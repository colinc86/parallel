package parallel

// ControllerConfiguration types contain the information necessary to calculate
// values from a PID controller.
type ControllerConfiguration struct {
	// The proportional term coefficient.
	Kp             float64

	// The integral term coefficient.
	Ki             float64

	// The derivative term coefficient.
	Kd             float64

	// The error function response.
	ErrorResponse  float64

	// The output signal response.
	OutputResponse float64
}

// NewControllerConfiguration creates and returns a new controller
// configuration.
func NewControllerConfiguration(kp float64, ki float64, kd float64, errorResponse float64, outputResponse float64) *ControllerConfiguration {
	return &ControllerConfiguration{
		Kp:             kp,
		Ki:             ki,
		Kd:             kd,
		ErrorResponse:  errorResponse,
		OutputResponse: outputResponse,
	}
}
