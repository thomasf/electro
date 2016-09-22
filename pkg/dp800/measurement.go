package dp800

import "fmt"

// Measurement .
type Measurement struct {
	Channel Channel
	Voltage float64
	Current float64
	Power   float64
}

func (m Measurement) String() string {
	return fmt.Sprintf("%s: %fv %fA %f...", m.Channel, m.Voltage, m.Current, m.Power)
}
