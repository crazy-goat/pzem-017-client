package main

import (
	"fmt"
	"github.com/goburrow/modbus"
	"time"
)

func printConfig(config Pzem017config)  {
	fmt.Println("Settings:")
	fmt.Printf(" * Modbus-RTU address:  %d\n", config.Address)
	fmt.Printf(" * High voltage alarm:  %.2f V\n", config.HighVoltageAlarm)
	fmt.Printf(" * Low voltage alarm:   %.2f V\n", config.LowVoltageAlarm)
	fmt.Printf(" * The current range:   %d A\n", config.Current)
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

func closePort(handler *modbus.RTUClientHandler) {
	err := handler.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
}