package ds1000z

import (
	"fmt"

	"github.com/thomasf/electro/pkg/lxi"
)

// DS1054z .
type DS1000Z struct {
	*lxi.Conn
}

func (d *DS1000Z) Connect() error {
	err := d.Conn.Connect()
	if err != nil {
		return err
	}
	i, err := d.IDN()

	if !(i.Model == "DS1054Z" || i.Model == "DS1104Z") {
		return fmt.Errorf("Epected model DS1054z, found %s", i.Model)
	}
	return nil
}
