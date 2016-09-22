package ut61usb

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zserge/hid"
)

func read(device hid.Device, ch chan []byte) error {
	if err := device.Open(); err != nil {
		log.Println("Open error: ", err)
		return err
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
		return err
	} else {
		log.Println("HID report", hex.EncodeToString(report))
	}

	var i int
	msg := make([]byte, 14)
loop:
	for {
		// fmt.Print(" read ")
		time.Sleep(5 * time.Millisecond)
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
				ch <- msg
				i = 0
			}
		} else {
			return err
		}
	}
	return nil
}

func ReadAll(ID string, recv chan []byte) error {
	var err error
	hid.UsbWalk(func(device hid.Device) {
		if err != nil {
			return
		}

		info := device.Info()
		id := fmt.Sprintf("%04x:%04x:%04x:%02x", info.Vendor, info.Product, info.Revision, info.Interface)
		if id != ID {
			return
		}
		err = read(device, recv)

	})
	return err
}

func OldMain() {
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

}
