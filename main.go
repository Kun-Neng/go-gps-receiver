package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

func main() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	} else {
		for _, port := range ports {
			fmt.Printf("Found port: %v\n", port)
		}
	}

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	serialPort, err := serial.Open("COM7", mode)
	if err != nil {
		log.Fatal(err)
	}

	receiveFromCom(serialPort)
}

func receiveFromCom(serialPort serial.Port) {
	buff := make([]byte, 512)
	for {
		time.Sleep(time.Second)

		n, err := serialPort.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			fmt.Println("\nEOF")
			break
		}
		// fmt.Printf("%v", string(buff[:n]))

		parseGPSInfo(string(buff[:n]))
	}
}

func parseGPSInfo(gpsInfo string) {
	strLineSlice := strings.Split(gpsInfo, "\n")
	if 0 == len(strLineSlice) {
		return
	}

	for _, oneLine := range strLineSlice {
		if 0 == len(oneLine) {
			continue
		}
		if '$' != oneLine[0] {
			continue
		}
		if !strings.Contains(oneLine, "*") {
			continue
		}
		if !strings.Contains(oneLine, "N") && !strings.Contains(oneLine, "S") {
			continue
		}
		if !strings.Contains(oneLine, "E") && !strings.Contains(oneLine, "W") {
			continue
		}

		if strings.Contains(oneLine, "GNGGA") {
			fmt.Printf("%v", oneLine)
			break
		}
		if strings.Contains(oneLine, "GNRMC") {
			fmt.Printf("%v", oneLine)
			break
		}
	}
}
