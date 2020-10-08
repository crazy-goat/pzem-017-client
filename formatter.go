package main

import (
	"errors"
	"fmt"
)

type Formatter interface {
	format(data Pzem017data) string
}

type FormatTxt struct {
	eol string
}

func (data FormatTxt) format(pzem Pzem017data) string {
	return fmt.Sprintf(
		"Time: %s, Address: %d, Voltage: %.2f V, Current: %.2f A, Power: %.1f W, Energy: %.3f kWh "+data.eol,
		pzem.Timestamp.Format("2006-01-02 15:04:05.000000"),
		pzem.Address,
		pzem.Voltage,
		pzem.Current,
		pzem.Power,
		float32(pzem.Energy)/1000.0)
}

type FormatterFactory struct {
	formatters map[string]Formatter
}

func (data FormatterFactory) add (name string, formatter Formatter) {
	data.formatters[name] = formatter
}

func (data FormatterFactory) getByName (format string) (result Formatter, err error) {
	formatter, exists := data.formatters[format]
	if exists == true {
		return formatter , nil
	}
	return nil, errors.New("format not implemented: "+format)
}
