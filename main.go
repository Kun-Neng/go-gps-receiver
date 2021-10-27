package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

type GPSInfo struct {
	Longitude    string
	Latitude     string
	LonDirection string
	LatDirection string
	LonRadian    float64
	LatRadian    float64
	IsGPSNormal  bool
}

var (
	directionMap = map[string]string{"N": "北緯", "S": "南緯", "E": "東經", "W": "西經"}
	GPSObject    = &GPSInfo{IsGPSNormal: false}
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
	var parseReadyFlag bool = false

	strLineSlice := strings.Split(gpsInfo, "\n")
	if 0 == len(strLineSlice) {
		GPSObject.IsGPSNormal = false
		return
	}

	for _, oneLine := range strLineSlice {
		if 0 == len(oneLine) {
			continue
		}
		if '$' != oneLine[0] {
			// Start of sentence
			continue
		}
		if !strings.Contains(oneLine, "*") {
			// Checksum delimiter
			continue
		}
		if !strings.Contains(oneLine, "N") && !strings.Contains(oneLine, "S") {
			continue
		}
		if !strings.Contains(oneLine, "E") && !strings.Contains(oneLine, "W") {
			continue
		}

		if strings.Contains(oneLine, "GNGGA") {
			// fmt.Printf("%v", oneLine)
			if false == parseGNGGA(oneLine) {
				continue
			}
			parseReadyFlag = true
			break
		}
		if strings.Contains(oneLine, "GNRMC") {
			// fmt.Printf("%v", oneLine)
			parseReadyFlag = true
			break
		}
	}

	if true == parseReadyFlag {
		GPSObject.IsGPSNormal = true
		fmt.Println(GPSObject.Latitude, GPSObject.Longitude)
	} else {
		GPSObject.IsGPSNormal = false
	}
}

func parseGNGGA(gnggaInfo string) bool {
	strSlice := strings.Split(gnggaInfo, ",")
	if 3 > len(strSlice[2]) || 4 > len(strSlice[4]) {
		return false
	}

	GPSObject.LatDirection = strSlice[3] // N/S
	GPSObject.LonDirection = strSlice[5] // E/W
	GPSObject.Latitude = directionMap[strSlice[3]] + strSlice[2][:2] + "度" + strSlice[2][2:] + "分"
	GPSObject.Longitude = directionMap[strSlice[5]] + strSlice[4][:3] + "度" + strSlice[4][3:] + "分"

	return true
}
