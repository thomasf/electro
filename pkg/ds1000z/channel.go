package ds1000z

type Channel int

const (
	ChCur Channel = iota
	Ch1
	Ch2
	Ch3
	Ch4
)

var (
	chStrMap = map[Channel]string{
		Ch1: "CH1",
		Ch2: "CH2",
		Ch3: "CH3",
		Ch4: "CH4",
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
	case Ch4:
		return "Ch4"

	default:
		panic("Invalid value")
	}
}
