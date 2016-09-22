package ut61usb

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CheckSuspend() error {
	// 	for dat in /sys/bus/usb/devices/*;
	//         do
	//         if test -e $dat/manufacturer; then
	//         	grep "WCH.CN" $dat/manufacturer>/dev/null&& echo auto >${dat}/power/level&&echo 0 > ${dat}/power/autosuspend
	//         fi
	// done
	basePath := "/sys/bus/usb/devices/"
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
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
				return err
			}
			level := strings.TrimSuffix(string(leveldata), "\n")
			if level != "auto" {
				return fmt.Errorf("not level 0: %v", level)
			}

			autosuspenddata, err := ioutil.ReadFile(filepath.Join(basePath, file.Name(), "power/autosuspend"))
			if err != nil {
				return err
			}
			autoSuspend := strings.TrimSuffix(string(autosuspenddata), "\n")
			if autoSuspend != "0" {
				return fmt.Errorf("autosuspend not 0: %v", autoSuspend)
			}
			log.Println(string(autosuspenddata))
		}
		// log.Printf("%s: '%s'",file.Name(), string(data))

	}
	return nil
}
