package main

import (
	"fmt"
	"github.com/devfacet/gocmd"
	"github.com/goburrow/modbus"
)

func registerCommandReadConfig(flags Commands) {
	_, _ = gocmd.HandleFlag("ReadConfig", func(cmd *gocmd.Cmd, args []string) error {
		return configRead(flags.ReadConfig.Port, byte(flags.ReadConfig.Address))
	})
}

func configRead(port string, address byte) error {
	handler := getHandler(port, address)
	err := handler.Connect()
	if err != nil {
		return fmt.Errorf("error while connecting: %s", err.Error())
	}

	client := modbus.NewClient(handler)
	data, err := client.ReadHoldingRegisters(0, 4)
	if err != nil {
		return fmt.Errorf("error while reading config registers: %s", err.Error())
	}

	config := createPzem017ConfigFromBytes(data)
	printConfig(config)

	return nil
}