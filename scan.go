package main

import (
	"fmt"
	"github.com/devfacet/gocmd"
	"github.com/goburrow/modbus"
	"go.bug.st/serial.v1/enumerator"
	"log"
	"time"
)

func scanForSlaves(port string, timeout time.Duration) {
	found := 0

	if timeout <= 0 {
		timeout = time.Duration(100)
	}
	fmt.Printf("Connecting port: %s\n", port)
	fmt.Printf("Timeout: %.3f\n", float32(timeout*time.Millisecond)/float32(time.Second))

	for address := 1; address < 127; address++ {
		handler := getHandlerWithTimeout(port, byte(address), timeout)
		fmt.Printf("Address %02d: ", address)
		err := handler.Connect()
		if err != nil {
			fmt.Println("error while connecting \"" + err.Error() + "\". Exiting")
			return
		}

		client := modbus.NewClient(handler)
		data, err := client.ReadHoldingRegisters(0, 4)

		if err != nil {
			fmt.Print(err.Error() + "\r")
		} else {
			fmt.Print("device found, checking response: ")
			config := createPzem017ConfigFromBytes(data)
			if config.validate() == true {
				fmt.Println("Ok")
				printConfig(config)
				found++
			} else {
				fmt.Println("Bad response")
			}
		}
		closePort(handler)
	}

	fmt.Printf("Total slaves found: %d\n", found)
}

func printSerialList(UsbOnly bool) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
		return
	}
	for _, port := range ports {
		if !UsbOnly || (UsbOnly && port.IsUSB) {
			fmt.Printf("Found port: %s\n", port.Name)
		}

		if port.IsUSB {
			fmt.Printf("   USB ID        %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial    %s\n", port.SerialNumber)
		}
	}
}

func registerCommandScan(flags Commands) {
	_, _ = gocmd.HandleFlag("Scan", func(cmd *gocmd.Cmd, args []string) error {
		scanForSlaves(flags.Scan.Port, time.Duration(flags.Scan.Timeout))
		return nil
	})
}

func registerCommandList(flags Commands) {
	_, _ = gocmd.HandleFlag("List", func(cmd *gocmd.Cmd, args []string) error {
		printSerialList(flags.List.UsbOnly)
		return nil
	})
}