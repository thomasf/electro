package ut61

import (
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// measurement .
type Measurement struct {
	Unit     string
	Value    float64
	Overload bool   // out of range
	Range    string // auto or manual
	ACDC     string
	Mode     string
	Hold     bool
}

func (m Measurement) String() string {
	var l []string
	l = append(l, fmt.Sprintf("%g %s", m.Value, m.Unit))
	if m.ACDC != "" {
		l = append(l, m.ACDC)
	}
	if m.Hold {
		l = append(l, "Hold")
	}
	if m.Mode != "" {
		l = append(l, m.Mode)

	}
	return strings.Join(l, " ")

}

func Parse(buf []byte) (Measurement, error) {

	var m Measurement

	// 0:     +/-
	// 1-4:   Digits (but see below)
	// 5:     Space
	// 6:     Precision
	// 7-8:   Flags
	// 9:     Prefix and special flags
	// 10:    Unit
	// 11:    Relative measurement integer
	// 12-13: CRLF
	// spew.Dump(buf)
	// spew.Dump(buf[7:12])
	// log.Printf("%b", buf[7:12])

	tt := binary.BigEndian.Uint32(buf[7:12])

	units := map[bits7t12]string{
		maskUnitPercent:   "%",
		maskUnitVolt:      "V",
		maskUnitAmpere:    "A",
		maskUnitOhm:       "Ω",
		maskUnitHz:        "Hz",
		maskUnitFarad:     "F",
		maskUnitCelsius:   "℃",
		maskUnitFarenheit: "℉",
	}

	for k, v := range units {
		if tt&uint32(k) > 1 {
			m.Unit = v
		}
	}

	var multiplier float64
	multipliers := map[bits7t12]float64{
		maskMultiplierNano:  1e-9,
		maskMultiplierMicro: 1e-6,
		maskMultiplierMilli: 1e-3,
		maskMultiplierKilo:  1e3,
		maskMultiplierMega:  1e6,
	}

	for k, v := range multipliers {
		if tt&uint32(k) > 1 {
			multiplier = v
		}
	}

	acdc := map[bits7t12]string{
		maskAC: "AC",
		maskDC: "DC",
	}
	for k, v := range acdc {
		if tt&uint32(k) > 1 {
			m.ACDC = v
		}
	}

	// TODO: Verufy that that these are mutually exclusive
	modes := map[bits7t12]string{
		maskModeMax:   "Max",
		maskModeMin:   "Min",
		maskModeRel:   "Rel",
		maskModeDiode: "Diode",
		maskModeBeep:  "Beep",
	}
	for k, v := range modes {
		if tt&uint32(k) > 1 {
			m.Mode = v
		}
	}
	m.Hold = tt&uint32(maskHold) > 1

	var diode, buzzer, percent bool

	switch buf[11] & 4 {
	case 0x00:
		// nothing
	case 0x02:
		percent = true
	case 0x04:
		diode = true
	case 0x08:
		buzzer = true
	}

	_ = diode && buzzer && percent

	valueStr := string(buf[1:5])
	var value int64
	if valueStr == "?0:?" {
		m.Overload = true
	} else {
		var err error
		value, err = strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return Measurement{}, err
		}
		if buf[0] == '-' {
			value = -value
		}

		m.Value = float64(value)

		if multiplier != 0 {
			m.Value = m.Value * multiplier
		}

		divisorPos, err := strconv.ParseInt(string(buf[6]), 10, 8)
		if err != nil {
			return Measurement{}, err
		}
		var divisor float64 = 1
		switch divisorPos {
		case 4:
			divisor = 10
		case 2:
			divisor = 100
		case 1:
			divisor = 1000
		}
		log.Println(
			"valueStr", valueStr,
			"mult", multiplier,
			"divisor", divisor,
		)
		m.Value = m.Value * (1.0 / divisor)

	}

	return m, nil

}

type bits7t12 uint32

const (
	maskUnitFarenheit   bits7t12 = 1 << iota //31
	maskUnitCelsius                          //30
	maskUnitFarad                            //29
	maskUnitHz                               //28
	_                                        //27
	maskUnitOhm                              //26
	maskUnitAmpere                           //25
	maskUnitVolt                             //24
	_                                        //23
	maskUnitPercent                          //22
	maskModeDiode                            //21
	maskModeBeep                             //20
	maskMultiplierMega                       //19
	maskMultiplierKilo                       //18
	maskMultiplierMilli                      //17
	maskMultiplierMicro                      //16
	_                                        //15
	maskMultiplierNano                       //14
	maskLowbat                               //13
	_                                        //12
	maskModeMin                              //11
	maskModeMax                              //10
	_                                        //9
	_                                        //8
	maskBG                                   //7
	maskHold                                 //6
	maskModeRel                              //5
	maskAC                                   //4
	maskDC                                   //3
	maskAuto                                 //2
	_                                        //1
)
