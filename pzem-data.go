package main

type Pzem017data struct {
	Voltage     float32
	Current     float32
	Power       float32
	Energy      int
	HighVoltage bool
	LowVoltage  bool
}

func CreatePzem017FromBytes(input []byte) Pzem017data {
	var data Pzem017data
	data.Voltage = float32(int(input[0])<<8+int(input[1])) / 100.0
	data.Current = float32(int(input[2])<<8+int(input[3])) / 100.0
	data.Power = float32(int(input[6])<<24+int(input[7])<<16+int(input[4])<<8+int(input[5])) / 10.0
	data.Energy = int(input[10])<<24 + int(input[11])<<16 + int(input[8])<<8 + int(input[9])
	data.HighVoltage = input[12] == 255 && input[13] == 255
	data.LowVoltage = input[14] == 255 && input[15] == 255

	return data
}
