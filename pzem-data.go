package main

import (
	"math"
	"time"
)

type Pzem017data struct {
	Voltage     float32
	Current     float32
	Power       float32
	Energy      int
	HighVoltage bool
	LowVoltage  bool
	Address		byte
	Timestamp   time.Time
}

type Pzem017config struct {
	HighVoltageAlarm float64
	LowVoltageAlarm float64
	Address byte
	Current int
}

func createPzem017ConfigFromBytes(input []byte) Pzem017config {
	data := Pzem017config{}
	data.HighVoltageAlarm = float64(int(input[0])<<8+int(input[1])) / 100.0
	data.LowVoltageAlarm = float64(int(input[2])<<8+int(input[3])) / 100.0
	data.Address = input[5]

	switch input[7] {
	case 0:
		data.Current = 100
	case 1:
		data.Current = 50
	case 2:
		data.Current = 200
	case 3:
		data.Current = 300
	default:
		data.Current = -1
	}
	
	return data
}

func (data Pzem017config) validate() bool {
	return data.Address <= 127 && data.Current > 0
}

func (data Pzem017data) validate() bool {
	return math.Abs(float64(data.Power-(data.Current*data.Voltage))) < 1.0
}

func CreatePzem017FromBytes(input []byte) Pzem017data {
	var data Pzem017data
	data.Voltage = float32(int(input[0])<<8+int(input[1])) / 100.0
	data.Current = float32(int(input[2])<<8+int(input[3])) / 100.0
	data.Power = float32(int(input[6])<<24+int(input[7])<<16+int(input[4])<<8+int(input[5])) / 10.0
	data.Energy = int(input[10])<<24 + int(input[11])<<16 + int(input[8])<<8 + int(input[9])
	data.HighVoltage = input[12] == 255 && input[13] == 255
	data.LowVoltage = input[14] == 255 && input[15] == 255
	data.Timestamp = time.Now()
	return data
}
