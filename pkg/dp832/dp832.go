package dp832

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/thomasf/electro/pkg/lxi"
)

// DP832 .
type DP832 struct {
	*lxi.Conn
}

func (d *DP832) Connect() error {
	err := d.Conn.Connect()
	if err != nil {
		return err
	}
	i, err := d.IDN()

	if i.Model != "DP832" {
		return fmt.Errorf("Epected model DP832, found %s", i.Model)
	}
	return nil
}

func (d *DP832) Measure(ch Channel) (Measurement, error) {
	var m Measurement
	cmd := fmt.Sprintf("MEAS:ALL? %s", chStrMap[ch])

	s, err := d.Command(cmd)
	if err != nil {
		return m, err
	}

	a := strings.Split(s, ",")

	{
		f, err := strconv.ParseFloat(a[0], 64)
		if err != nil {
			return m, err
		}
		m.Current = f
	}

	{
		f, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			return m, err
		}
		m.Voltage = f
	}

	{
		f, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			return m, err
		}
		m.Power = f
	}
	m.Channel = ch

	return m, nil
}
