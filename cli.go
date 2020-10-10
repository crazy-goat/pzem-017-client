package main

import (
	"github.com/devfacet/gocmd"
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
	ReadConfig struct{
		Port     string `short:"p" long:"port" required:"true" description:"Serial port"`
		Address  int    `short:"a" long:"address" required:"true" description:"Slave address"`
	} `command:"config-get" description:"Get PZEM-107 config"`
	Formats struct {} `command:"show-formats" description:"Show available output formats"`
}

func registerCli() {
	flags := Commands{}

	registerCommandRead(flags)
	registerCommandList(flags)
	registerCommandScan(flags)
	registerCommandFormats(flags)
	registerCommandReset(flags)
	registerCommandReadConfig(flags)

	// Init the app
	_, _ = gocmd.New(gocmd.Options{
		Name:        "pzem-017-client",
		Version:     "1.0.0",
		Description: "Pzem-017 power metter reader",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})
}