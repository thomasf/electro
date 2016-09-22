package dp800

type Channel int

const (
	ChCur Channel = iota
	Ch1
	Ch2
	Ch3
)

var (
	chStrMap = map[Channel]string{
		ChCur: "",
		Ch1:   "CH1",
		Ch2:   "CH2",
		Ch3:   "CH3",
	}
)

func (c Channel) String() string {
	switch c {
	case ChCur:
		return "ChC"
	case Ch1:
		return "Ch1"
	case Ch2:
		return "Ch2"
	case Ch3:
		return "Ch3"
	default:
		panic("Invalid value")
	}
}

type chRange struct {
	VoltageMin float64
	VoltageMax float64
	CurrentMin float64
	CurrentMax float64
}

var chRanges = map[Channel]chRange{
	Ch1: {
		VoltageMin: 00.000,
		VoltageMax: 32.000,
		CurrentMin: 00.000,
		CurrentMax: 3.200,
	},
	Ch2: {
		VoltageMin: 00.000,
		VoltageMax: 32.000,
		CurrentMin: 00.000,
		CurrentMax: 3.200,
	},
	Ch3: {
		VoltageMin: 00.000,
		VoltageMax: 5.300,
		CurrentMin: 00.000,
		CurrentMax: 3.200,
	},
}
