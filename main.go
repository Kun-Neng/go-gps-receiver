package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Kun-Neng/go-gps-receiver/v0.1.0/publisher"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
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

	Publisher = publisher.GetInstance()
)

func main() {
	log.Println("Author: Kun-Neng Hung")
	log.Println("OS:", runtime.GOOS)

	cmd := exec.Command("whoami")
	user, _ := cmd.Output()
	log.Printf("User name: %s", user)

	cmd = exec.Command("ls", "-al")
	output, _ := cmd.Output()
	log.Printf("%s", output)

	// Publisher.Listen("tcp://*:5555")
	Publisher.ListenLocal()
	defer Publisher.Quit()

	for {
		Ports = []string{}

		if ok := ScanPorts(); ok {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter port name: ")
			port, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			port = trimInput(runtime.GOOS, port)

			hasPort := contains(Ports, port)
			if port == "" || hasPort == false {
				log.Printf("Port %s doesn't exist!\n", port)
				continue
			}

			InitDevice(port, defaultBaudRate)

			log.Println("Publishing data every 250 milliseconds ...")
			Publisher.Start()
			go Receive(1000)
			go Send()

			for {
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func ScanPorts() bool {
	// ports, err := serial.GetPortsList()
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
		return false
	}

	if len(ports) == 0 {
		log.Println("No serial ports found!")
		return false
	}

	for _, port := range ports {
		Ports = append(Ports, port.Name)
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

	log.Println("Init Device", ComObject.PortName, ComObject.BaudRate)
}

func Receive(millisecond int64) {
	ComObject.ReceiveFromCom(millisecond)
}

func Send() {
	for {
		data := <-ComObject.DataChannel
		// fmt.Printf("%v", data)
		Publisher.Update(data)
	}
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

func trimInput(os string, str string) string {
	if os == "windows" {
		return strings.TrimRight(str, "\r\n")
	} else {
		return strings.TrimRight(str, "\n")
	}
}

func contains(strArray []string, str string) bool {
	for _, v := range strArray {
		if v == str {
			return true
		}
	}
	return false
}
