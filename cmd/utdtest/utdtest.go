package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zserge/hid"
)

func chechSuspend() {
	// 	for dat in /sys/bus/usb/devices/*;
	//         do
	//         if test -e $dat/manufacturer; then
	//         	grep "WCH.CN" $dat/manufacturer>/dev/null&& echo auto >${dat}/power/level&&echo 0 > ${dat}/power/autosuspend
	//         fi
	// done
	basePath := "/sys/bus/usb/devices/"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}

devices:
	for _, file := range files {

		// fmt.Println(file.Name())
		filename := filepath.Join(basePath, file.Name(), "manufacturer")
		data, err := ioutil.ReadFile(filename)
		if os.IsNotExist(err) {
			continue devices
		}
		if err != nil {
			log.Fatalln(err)
		}

		if strings.HasPrefix(string(data), "WCH.CN") {
			log.Println("FOUND", file)
			leveldata, err := ioutil.ReadFile(filepath.Join(basePath, file.Name(), "power/level"))
			if err != nil {
				log.Fatalln(err)
			}
			level := strings.TrimSuffix(string(leveldata), "\n")
			if level != "auto" {
				log.Fatalln("not level 0", level)
			}

			autosuspenddata, err := ioutil.ReadFile(filepath.Join(basePath, file.Name(), "power/autosuspend"))
			if err != nil {
				log.Fatalln(err)
			}
			autoSuspend := strings.TrimSuffix(string(autosuspenddata), "\n")
			if autoSuspend != "0" {
				log.Fatalln("autosuspend not 0", autoSuspend)
			}
			log.Println(string(autosuspenddata))
		}
		// log.Printf("%s: '%s'",file.Name(), string(data))

	}
}

func shell(device hid.Device) {
	if err := device.Open(); err != nil {
		log.Println("Open error: ", err)
		return
	}
	defer device.Close()

	var bps uint32 = 19230
	init := []byte{
		0,
		byte(bps),
		byte(bps >> 8),
		byte(bps >> 16),
		byte(bps >> 24),
		0x03,
	}
	if err := device.SetReport(0, init); err != nil {
		log.Fatalln(err)
	}
	if report, err := device.HIDReport(); err != nil {
		log.Println("HID report error:", err)
		return
	} else {
		log.Println("HID report", hex.EncodeToString(report))
	}

	go func() {
		var i int
		msg := make([]byte, 14)
	loop:
		for {
			// fmt.Print(" read ")
			time.Sleep(10 * time.Millisecond)
			if buf, err := device.Read(8, 1*time.Second); err == nil {
				length := buf[0] & 7
				if length == 0 {
					// fmt.Print(" 0len ")
					continue loop
				}
				if len(buf) < 1+int(length) {
					log.Println("invalid length")
					continue loop
				}
				for _, v := range buf[1 : 1+length] {
					if v == '\n' && i != 13 {
						log.Println("got newline as char ", i)
						i = 0
						continue loop
					}
					msg[i] = v
					i++
				}
				if i >= 14 {
					m, err := parseMsg(msg)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(m)
					// spew.Dump(m)
					// _ = m
					// log.Println(hex.EncodeToString(msg), ":", string(msg[0:12]))
					i = 0
				}
			} else {

				log.Fatal(err)
			}
		}
	}()

	for {
	}

}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("USAGE:")
		fmt.Printf("  %s              list USB HID devices\n", os.Args[0])
		fmt.Printf("  %s <id>         open USB HID device shell for the given input report size\n", os.Args[0])
		fmt.Printf("  %s -h|--help    show this help\n", os.Args[0])
		fmt.Println()
		return
	}

	chechSuspend()

	// Without arguments - enumerate all HID devices
	if len(os.Args) == 1 {
		found := false
		hid.UsbWalk(func(device hid.Device) {
			info := device.Info()
			fmt.Printf("%04x:%04x:%04x:%02x\n", info.Vendor, info.Product, info.Revision, info.Interface)
			found = true
		})
		if !found {
			fmt.Println("No USB HID devices found")
		}
		return
	}

	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		id := fmt.Sprintf("%04x:%04x:%04x:%02x", info.Vendor, info.Product, info.Revision, info.Interface)
		if id != os.Args[1] {
			return
		}
		shell(device)
	})
}

// measurement .
type measurement struct {
	Unit     string
	Value    float64
	Overload bool   // out of range
	Range    string // auto or manual
	ACDC     string
	Mode     string
	Hold     bool
}

func (m measurement) String() string {
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

func parseMsg(buf []byte) (measurement, error) {

	var m measurement

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
			return measurement{}, err
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
			return measurement{}, err
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
