package main

import (
	"errors"
	"fmt"
	"github.com/devfacet/gocmd"
	"github.com/goburrow/modbus"
	"time"
)

func readData(port string, address byte, formatter Formatter, interval time.Duration) {
	handler := getHandler(port, address)
	defer closePort(handler)
	_ = handler.Connect()

	client := modbus.NewClient(handler)

	for {
		start := time.Now()

		data, err := readSingleMeasurement(client)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		data.Address = address
		fmt.Print(formatter.format(data))

		sleep := interval - time.Now().Sub(start)
		if sleep < 0 {
			sleep = 0
		}

		time.Sleep(sleep)
	}
}

func readSingleMeasurement(client modbus.Client) (data Pzem017data, err error) {
	results, err := client.ReadInputRegisters(0, 8)
	if err != nil {
		return Pzem017data{}, err
	}

	data = CreatePzem017FromBytes(results)
	if data.validate() == false {
		return Pzem017data{}, errors.New("Invalid data")
	}

	return data, nil
}

func registerCommandRead(flags *Commands) {
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
}