package main

import (
	"fmt"
	"log"
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
		fmt.Printf("%v", string(buff[:n]))
	}
}
