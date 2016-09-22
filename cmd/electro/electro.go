package main

import (
	"flag"
	"log"
	"os"
	"time"

	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/thomasf/electro/pkg/dp800"
	"github.com/thomasf/electro/pkg/ds1000z"
	"github.com/thomasf/electro/pkg/lxi"
)

func main() {
	var addrFlag = flag.String("addr", "192.168.0.123:5555", "tcp hostport")
	flag.Parse()

	c := &lxi.Conn{
		Addr: *addrFlag,
	}
	err := c.Connect()
	if err != nil {
		panic(err)
	}
	i, err := c.IDN()
	if err != nil {
		panic(err)
	}
	spew.Dump(i)
	switch i.Model {
	case "DP832":
		dp832Test(c)
	case "DS1054Z", "DS1104Z":
		ds1000Test(c)
	default:
		log.Println("unsupported device", i.Model)
		os.Exit(1)
	}

}

func dp832Test(c *lxi.Conn) {
	dp := dp800.DP800{Conn: c}
	err := dp.Connect()
	if err != nil {
		log.Fatal(err)
	}

	for range time.Tick(100 * time.Millisecond) {
		for _, CH := range []dp800.Channel{dp800.Ch1, dp800.Ch2, dp800.Ch3} {
			m, err := dp.Measure(CH)
			if err != nil {
				panic(err)
			}
			log.Printf("%v", m)
		}
	}

}

func memdepth(ds *ds1000z.DS1000Z) float64 {
	//	 Define number of horizontal grid divisions for DS1054Z
	const hGrid = 12

	mdep, err := ds.Command("ACQ:MDEP?")
	if err != nil {
		panic(err)
	}

	if mdep == "AUTO" {
		srat, err := ds.Command("ACQ:SRAT?")
		if err != nil {
			panic(err)
		}
		sampleRate, err := strconv.ParseFloat(srat, 64)
		if err != nil {
			panic(err)
		}

		// # TIMebase[:MAIN]:SCALe
		tim, err := ds.Command("TIM:SCAL?")
		if err != nil {
			panic(err)
		}
		timeBase, err := strconv.ParseFloat(tim, 64)
		if err != nil {
			panic(err)
		}

		memoryDepth := hGrid * timeBase * sampleRate
		return memoryDepth

	}

	memoryDepth, err := strconv.ParseFloat(mdep, 64)
	if err != nil {
		panic(err)
	}

	return memoryDepth

}

func ds1000Test(c *lxi.Conn) {
	ds := &ds1000z.DS1000Z{Conn: c}
	err := ds.Connect()
	if err != nil {
		panic(err)
	}

	// var channels []string
channels:
	for _, ch := range []string{"CHAN1", "chan2", "chan3", "chan4", "math"} {
		resp, err := ds.Command(":"+ch, "display?")
		if err != nil {
			panic(err)
		}
		if resp == "0" {
			continue channels
		}
		log.Println(ch)
		log.Println("memory depth", memdepth(ds))
		//# Set WAVE parameters
		//  tn.write("waveform:source " + channel)
		//   time.sleep(1)

		//    tn.write("waveform:form asc")
		//   time.sleep(1)

		//   # Maximum - only displayed data when osc. in RUN mode, or full memory data when STOPed
		//  tn.write("waveform:mode max")
		//  time.sleep(1)
	}

}
