package main

import (
	"fmt"
	"github.com/devfacet/gocmd"
	"github.com/goburrow/modbus"
	"go.bug.st/serial.v1/enumerator"
	"log"
	"time"
)

type Pzem017data struct {
	Voltage     float32
	Current     float32
	Power       float32
	Energy      int
	HighVoltage bool
	LowVoltage  bool
}

func (p *Pzem017data) fromBytes(data []byte) {
	p.Voltage = float32(int(data[0])<<8+int(data[1])) / 100.0
	p.Current = float32(int(data[2])<<8+int(data[3])) / 100.0
	p.Power = float32(int(data[6])<<24+int(data[7])<<16+int(data[4])<<8+int(data[5])) / 10.0
	p.Energy = int(data[10])<<24 + int(data[11])<<16 + int(data[8])<<8 + int(data[9])
	p.HighVoltage = data[12] == 255 && data[13] == 255
	p.LowVoltage = data[14] == 255 && data[15] == 255
}

func closePort(handler *modbus.RTUClientHandler) {
	err := handler.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
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

func getHandler(port string, slaveId byte) *modbus.RTUClientHandler {
	return getHandlerWithTimeout(port, slaveId, 5000)
}

func getHandlerWithTimeout(port string, slaveId byte, timeout time.Duration) *modbus.RTUClientHandler {

	handler := modbus.NewRTUClientHandler(port)
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = slaveId
	handler.Timeout = timeout * time.Millisecond

	return handler
}

func readData() {
	handler := getHandler("/dev/ttyUSB01", 1)
	defer closePort(handler)
	_ = handler.Connect()

	client := modbus.NewClient(handler)

	for {
		results, err := client.ReadInputRegisters(0, 8)
		t := Pzem017data{}
		t.fromBytes(results)

		fmt.Printf(
			"Voltage: %.2f V, Current: %.2f A, Power: %.1f W, Energy: %.3f kWh \r",
			t.Voltage,
			t.Current,
			t.Power,
			float32(t.Energy)/1000.0)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	flags := struct {
		List struct {
			UsbOnly bool `short:"u" long:"usb" description:"Display USB only ports"`
		} `command:"list" description:"Show list of available serial ports"`
		Scan struct {
			Port    string `short:"p" long:"port" required:"true" description:"Serial port"`
			Timeout int64  `short:"t" long:"timeout"  description:"Timeout in milliseconds"`
		} `command:"scan" description:"Scan for modbus slaves"`
		Read struct {
			Port    string `short:"p" long:"port" required:"true" description:"Serial port"`
			Address string `short:"a" long:"address" required:"true" description:"Slave address, if more use 1,2,3"`
		} `command:"read" description:"Read data from pzem-017 slaves"`
	}{}

	_, _ = gocmd.HandleFlag("List", func(cmd *gocmd.Cmd, args []string) error {
		printSerialList(flags.List.UsbOnly)
		return nil
	})

	_, _ = gocmd.HandleFlag("Scan", func(cmd *gocmd.Cmd, args []string) error {
		scanForSlaves(flags.Scan.Port, time.Duration(flags.Scan.Timeout))
		return nil
	})

	// Init the app
	_, _ = gocmd.New(gocmd.Options{
		Name:        "pzem-017-client",
		Version:     "1.0.0",
		Description: "Pzem-017 power metter reader",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})
}

func scanForSlaves(port string, timeout time.Duration) {
	found := 0

	if timeout <= 0 {
		timeout = time.Second
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
		_, err = client.ReadInputRegisters(0, 8)

		if err != nil {
			fmt.Print(err.Error() + "\r")
		} else {
			fmt.Println("Ok")
			found++
		}
		_ = handler.Close()
	}

	fmt.Printf("Total slaves found: %d\r", found)
}
