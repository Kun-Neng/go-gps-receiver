package main

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

type Com struct {
	PortName     string
	BaudRate     int
	SerialPort   *serial.Port
	DataChannel  chan string
	CloseChannel chan bool
	IsComNormal  bool
}

var (
	Ports []string

	defaultPortName = "COM7"
	defaultBaudRate = 115200
	localTimeZone   = "Asia/Taipei"
	datetimeLayout  = "2006-01-02 15:04:05"

	ComObject = &Com{
		PortName:     defaultPortName,
		BaudRate:     defaultBaudRate,
		DataChannel:  make(chan string),
		CloseChannel: make(chan bool, 1),
		IsComNormal:  false,
	}
)

func main() {
	if ok := ScanPorts(); ok {
		InitDevice(Ports[0], 115200)
		go Receive(1000)
		go func() {
			for {
				gpsInfo := <-ComObject.DataChannel
				fmt.Printf("%v", gpsInfo)
			}
		}()

		for {
		}
	}
}

func ScanPorts() bool {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
		return false
	}

	if len(ports) == 0 {
		log.Println("No serial ports found!")
		return false
	}

	for _, port := range ports {
		Ports = append(Ports, port)
	}
	log.Printf("Found port: %v\n", Ports)

	return true
}

func InitDevice(port string, baudrate int) {
	if port == "" && baudrate == 0 {
		if true == ComObject.GetPortName() {
			defaultPortName = ComObject.PortName
		}
	} else {
		ComObject.PortName = port
		ComObject.BaudRate = baudrate
	}

	err := ComObject.OpenComPort()
	if err == nil {
		ComObject.IsComNormal = true
	}

	log.Println("Init Device")
}

func Receive(millisecond int64) {
	ComObject.ReceiveFromCom(millisecond)
}

func CloseDevice() {
	ComObject.Close()
}

func (com *Com) GetPortName() bool {
	if 0 < len(Ports) {
		com.PortName = Ports[0]
		return true
	}

	return false
}

func (com *Com) OpenComPort() (err error) {
	mode := &serial.Mode{
		BaudRate: com.BaudRate,
	}

	serialPort, err := serial.Open(com.PortName, mode)
	if err != nil {
		log.Fatal(err)
		return err
	}

	com.SerialPort = &serialPort

	return nil
}

func (com *Com) ReceiveFromCom(millisecond int64) {
	defer com.Close()

	buff := make([]byte, 512)
	for {
		time.Sleep(time.Duration(millisecond) * time.Millisecond)

		n, err := (*com.SerialPort).Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}

		if n == 0 {
			log.Println("\nEOF")
			break
		}

		com.DataChannel <- string(buff[:n])
	}
}

func (com *Com) Close() {
	(*com.SerialPort).Close()
	close(com.DataChannel)
	com.CloseChannel <- true
}
