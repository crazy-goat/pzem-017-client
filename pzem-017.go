package main

import (
	"encoding/binary"
	"fmt"
	"github.com/devfacet/gocmd"
	"github.com/goburrow/modbus"
	"go.bug.st/serial.v1/enumerator"
	"log"
	"time"
)

type Commands struct {
	List struct {
		UsbOnly bool `short:"u" long:"usb" description:"Display USB only ports"`
	} `command:"list" description:"Show list of available serial ports"`
	Scan struct {
		Port    string `short:"p" long:"port" required:"true" description:"Serial port"`
		Timeout int64  `short:"t" long:"timeout"  description:"Timeout in milliseconds"`
	} `command:"scan" description:"Scan for modbus slaves"`
	Read struct {
		Port     string `short:"p" long:"port" required:"true" description:"Serial port"`
		Address  int    `short:"a" long:"address" required:"true" description:"Slave address"`
		Format   string `short:"f" long:"format" description:"Output format. Default txt"`
		Interval int    `short:"i" long:"interval" description:"Read interval in millisecondsr"`
	} `command:"read" description:"Read data from pzem-017 slaves"`
	Reset struct{
		Port     string `short:"p" long:"port" required:"true" description:"Serial port"`
		Address  int    `short:"a" long:"address" required:"true" description:"Slave address"`
	} `command:"reset" description:"Set energy counter to 0"`
	Formats struct {} `command:"show-formats" description:"Show available output formats"`
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

func readData(port string, address byte, formatter Formatter, interval time.Duration) {
	handler := getHandler(port, address)
	defer closePort(handler)
	_ = handler.Connect()

	client := modbus.NewClient(handler)

	for {
		results, err := client.ReadInputRegisters(0, 8)
		if err != nil {
			fmt.Println(err)
			return
		}
		data := CreatePzem017FromBytes(results, address)
		if data.validate() == false{
			fmt.Println("Invalid data")
		} else {
			fmt.Printf(formatter.format(data))
		}


		time.Sleep(interval)
	}
}

func getFormatFactory() FormatterFactory {
	factory := FormatterFactory{formatters: make(map[string]Formatter)}
	factory.add("txt", FormatTxt{eol: "\r"})
	factory.add("txt-newline", FormatTxt{eol: "\n"})
	return factory
}

func main() {
	flags := Commands{}

	_, _ = gocmd.HandleFlag("List", func(cmd *gocmd.Cmd, args []string) error {
		printSerialList(flags.List.UsbOnly)
		return nil
	})

	_, _ = gocmd.HandleFlag("Read", func(cmd *gocmd.Cmd, args []string) error {
		format := flags.Read.Format
		if format == "" {
			format = "txt"
		}

		formatter, err := getFormatFactory().getByName(format)

		if err != nil {
			fmt.Println(err.Error())
			return nil
		}

		readData(
			flags.Read.Port,
			byte(flags.Read.Address),
			formatter, time.Duration(flags.Read.Interval)*time.Millisecond)

		return nil
	})

	_, _ = gocmd.HandleFlag("Scan", func(cmd *gocmd.Cmd, args []string) error {
		scanForSlaves(flags.Scan.Port, time.Duration(flags.Scan.Timeout))
		return nil
	})

	_, _ = gocmd.HandleFlag("Formats", func(cmd *gocmd.Cmd, args []string) error {
		fmt.Println("Available formats:")
		for key, _ := range getFormatFactory().formatters {
			fmt.Println(" * " + key)
		}
		return nil
	})

	_, _ = gocmd.HandleFlag("Reset", func(cmd *gocmd.Cmd, args []string) error {
		err := resetEnergy(flags.Reset.Port, byte(flags.Reset.Address))
		if err != nil {
			fmt.Println("Energy meter set to 0")
		}

		return err
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

func resetEnergy(port string, address byte) error {
	handler := getHandler(port, address)

	defer closePort(handler)
	err := handler.Connect()
	if err != nil {
		return nil
	}

	request := modbus.ProtocolDataUnit{
		FunctionCode: 0x42,
		Data:         dataBlock(uint16(address)),
	}
	response, err := send(handler, &request)

	if err != nil {
		return err
	}

	if len(response.Data) != 2 {
		return fmt.Errorf("modbus: response data size '%v' does not match expected '%v'", len(response.Data), 2)
	}

	respValue := binary.BigEndian.Uint16(response.Data)
	if uint16(address) != respValue {
		return  fmt.Errorf("modbus: response address '%v' does not match request '%v'", respValue, address)
	}

	return nil
}
func send(handler modbus.ClientHandler, request *modbus.ProtocolDataUnit) (response *modbus.ProtocolDataUnit, err error) {
	aduRequest, err := handler.Encode(request)
	if err != nil {
		return
	}
	aduResponse, err := handler.Send(aduRequest)
	if err != nil {
		return
	}
	if err = handler.Verify(aduRequest, aduResponse); err != nil {
		return
	}
	response, err = handler.Decode(aduResponse)
	if err != nil {
		return
	}
	// Check correct function code returned (exception)
	if response.FunctionCode != request.FunctionCode {
		err = fmt.Errorf("modbus: response with different code")
		return
	}
	if response.Data == nil || len(response.Data) == 0 {
		// Empty response
		err = fmt.Errorf("modbus: response data is empty")
		return
	}
	return
}

func dataBlock(value ...uint16) []byte {
	data := make([]byte, 2*len(value))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	return data
}
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
				fmt.Println("Settings:")
				fmt.Printf(" * Modbus-RTU address:  %d\n", config.Address)
				fmt.Printf(" * High voltage alarm:  %.2f V\n", config.HighVoltageAlarm)
				fmt.Printf(" * Low voltage alarm:   %.2f V\n", config.LowVoltageAlarm)
				fmt.Printf(" * The current range:   %d A\n", config.Current)
				found++
			} else {
				fmt.Println("Bad response")
			}
		}
		closePort(handler)
	}

	fmt.Printf("Total slaves found: %d\n", found)
}
