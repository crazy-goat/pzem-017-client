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
		"Voltage: %.2f V, Current: %.2f A, Power: %.1f W, Energy: %.3f kWh "+data.eol,
		pzem.Voltage,
		pzem.Current,
		pzem.Power,
		float32(pzem.Energy)/1000.0)
}

func formatterFactory (format string) (result Formatter, err error) {
	switch format {
		case "txt":
			return FormatTxt{eol: "\r"}, nil
		case "txt-newline":
			return FormatTxt{eol: "\n"}, nil
	default:
		return nil, errors.New("format not implemented: "+format)
	}
}
