package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Com struct {
	PortName     string
	BaudRate     int
	SerialPort   *serial.Port
	CloseChannel chan bool
	IsComNormal  bool
}

type GPSInfo struct {
	Longitude    string
	Latitude     string
	LonDirection string
	LatDirection string
	LonRadian    float64
	LatRadian    float64
	MessageFrom  string
	IsGPSNormal  bool
}

var (
	defaultPortName = "COM7"
	ComObject       = &Com{PortName: defaultPortName, BaudRate: 115200, CloseChannel: make(chan bool, 1), IsComNormal: false}
	localTimeZone   = "Asia/Taipei"
	timeLayout      = "2006-01-02 15:04:05"
	minLatLen       = 3
	minLonLen       = 4
	directionMap    = map[string]string{"N": "北緯", "S": "南緯", "E": "東經", "W": "西經"}
	GPSObject       = &GPSInfo{IsGPSNormal: false}
)

func main() {
	ComObject.PortName = defaultPortName
	if true == ComObject.GetPortName() {
		defaultPortName = ComObject.PortName
	}

	err := ComObject.OpenComPort()
	if err != nil {
		GPSObject.IsGPSNormal = false
	} else {
		ComObject.IsComNormal = true
	}

	localTime, err := time.LoadLocation(localTimeZone)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(time.Now().In(localTime).Format(timeLayout))

	ComObject.ReceiveFromCom()
}

func (this *Com) GetPortName() bool {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
		return false
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
		return false
	}

	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}

	this.PortName = ports[0]

	return true
}

func (this *Com) OpenComPort() (err error) {
	mode := &serial.Mode{
		BaudRate: this.BaudRate,
	}

	serialPort, err := serial.Open(this.PortName, mode)
	if err != nil {
		log.Fatal(err)
		return err
	}

	this.SerialPort = &serialPort

	return nil
}

func (this *Com) Close() {
	(*this.SerialPort).Close()
	this.CloseChannel <- true
}

func (this *Com) ReceiveFromCom() {
	defer this.Close()

	buff := make([]byte, 512)
	for {
		time.Sleep(time.Second)

		n, err := (*this.SerialPort).Read(buff)
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
			fmt.Println(parseDateTime(oneLine))
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
		fmt.Println(GPSObject.MessageFrom, GPSObject.Latitude, GPSObject.Longitude)
	} else {
		GPSObject.IsGPSNormal = false
	}
}

func parseDateTime(oneLineInfo string) string {
	strSlice := strings.Split(oneLineInfo, ",")

	timeToken := strSlice[1]
	// hour, _ := strconv.Atoi(timeToken[:2])
	// min, _ := strconv.Atoi(timeToken[2:4])
	// sec, _ := strconv.Atoi(timeToken[4:6])
	// msec, _ := strconv.Atoi(timeToken[7:9])
	// nsec := 1000 * msec

	dateToken := strSlice[9]
	// year, _ := strconv.Atoi("20" + dateToken[4:6])
	// month, _ := strconv.Atoi(dateToken[2:4])
	// timeMonth := time.Month(month)
	// day, _ := strconv.Atoi(dateToken[0:2])
	// fmt.Printf("%d年%d月%d日\n", year, month, day)

	// timeString := time.Date(year, timeMonth, day, hour, min, sec, nsec, time.UTC).String()

	location, err := time.LoadLocation(localTimeZone)
	if err != nil {
		log.Fatal(err)
	}

	timeStringArray := []string{"20", dateToken[4:6], "-", dateToken[2:4], "-", dateToken[0:2], " ", timeToken[:2], ":", timeToken[2:4], ":", timeToken[4:6]}
	timeString := strings.Join(timeStringArray, "")
	// fmt.Println(timeString)
	timeInUTC, _ := time.Parse(timeLayout, timeString)
	// fmt.Println(timeInUTC)

	timeInLocation := timeInUTC.In(location)
	// fmt.Println(timeInLocation)

	return timeInLocation.String()
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

	GPSObject.MessageFrom = strSlice[0]

	return true
}
