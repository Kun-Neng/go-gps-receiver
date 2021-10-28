package main

import (
	"fmt"
	"log"
	"strconv"
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
	messageFrom  string
	IsGPSNormal  bool
}

var (
	minLatLen    = 3
	minLonLen    = 4
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
	var parseLatLonReadyFlag bool = false

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

		if strings.Contains(oneLine, "GPRMC") || strings.Contains(oneLine, "GNRMC") {
			// fmt.Printf("%v", oneLine)
			parseDateTime(oneLine)
			if false == parseLatLon(oneLine, 3, 5) {
				continue
			}
			parseLatLonReadyFlag = true

			break
		}
		if strings.Contains(oneLine, "GPGGA") || strings.Contains(oneLine, "GNGGA") {
			// fmt.Printf("%v", oneLine)
			if false == parseLatLon(oneLine, 2, 4) {
				continue
			}
			parseLatLonReadyFlag = true

			break
		}
	}

	if true == parseLatLonReadyFlag {
		GPSObject.IsGPSNormal = true
		fmt.Println(GPSObject.messageFrom, GPSObject.Latitude, GPSObject.Longitude)
	} else {
		GPSObject.IsGPSNormal = false
	}
}

func parseDateTime(oneLineInfo string) {
	strSlice := strings.Split(oneLineInfo, ",")
	timeToken := strSlice[1]

	dateToken := strSlice[9]
	year := dateToken[4:6]
	month := dateToken[2:4]
	day := dateToken[0:2]

	fmt.Printf("20%s 年 %s 月 %s 日 %v\n", year, month, day, timeToken)
}

func parseLatLon(oneLineInfo string, iLat int, iLon int) bool {
	strSlice := strings.Split(oneLineInfo, ",")
	if len(strSlice[iLat]) < minLatLen || len(strSlice[iLon]) < minLonLen {
		return false
	}

	iLatDirection := iLat + 1
	iLonDirection := iLon + 1

	GPSObject.LatDirection = strSlice[iLatDirection] // N/S
	GPSObject.LonDirection = strSlice[iLonDirection] // E/W
	GPSObject.Latitude = directionMap[strSlice[iLatDirection]] + strSlice[iLat][:2] + "度" + strSlice[iLat][2:] + "分"
	GPSObject.Longitude = directionMap[strSlice[iLonDirection]] + strSlice[iLon][:3] + "度" + strSlice[iLon][3:] + "分"

	tmpIntPartLat, _ := strconv.ParseFloat(strSlice[iLat][:2], 32)
	tmpDecimalPartLat, _ := strconv.ParseFloat(strSlice[iLat][2:], 32)
	GPSObject.LatRadian = tmpIntPartLat + tmpDecimalPartLat/60

	tmpIntPartLon, _ := strconv.ParseFloat(strSlice[iLon][:3], 32)
	tmpDecimalPartLon, _ := strconv.ParseFloat(strSlice[iLon][3:], 32)
	GPSObject.LonRadian = tmpIntPartLon + tmpDecimalPartLon/60

	GPSObject.messageFrom = strSlice[0]

	return true
}
