package main

import (
	"encoding/json"
	"fmt"
	"github.com/devfacet/gocmd"
	"github.com/goburrow/modbus"
	"net/http"
)

type Pzem017json struct {
	Voltage     float32
	Current     float32
	Power       float32
	Energy      int
	HighVoltage bool
	LowVoltage  bool
	Name        string
}

type serveJson struct {
	Name    string
	Client  modbus.Client
	Address byte
}

func (data Pzem017data) createJson(name string) Pzem017json {
	jsonData := Pzem017json{}

	jsonData.Current = data.Current
	jsonData.Voltage = data.Voltage
	jsonData.Power = data.Power
	jsonData.Energy = data.Energy
	jsonData.HighVoltage = data.HighVoltage
	jsonData.LowVoltage = data.LowVoltage
	if name == "" {
		jsonData.Name = fmt.Sprintf("%d", data.Address)
	} else {
		jsonData.Name = name
	}

	return jsonData
}

func (di *serveJson) indexController(w http.ResponseWriter, req *http.Request) {
	data, err := readSingleMeasurement(di.Client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data.Address = di.Address
	response := data.createJson(di.Name)

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func registerCommandServer(flags *Commands) {
	_, _ = gocmd.HandleFlag("Http", func(cmd *gocmd.Cmd, args []string) error {
		handler := getHandler(flags.Http.Port, byte(flags.Http.Address))
		defer closePort(handler)
		_ = handler.Connect()
		controller := &serveJson{
			Name:   flags.Http.Name,
			Client: modbus.NewClient(handler),
			Address: byte(flags.Http.Address)}

		http.HandleFunc("/", controller.indexController)
		err := http.ListenAndServe(
			fmt.Sprintf(":%d", flags.Http.HttpPort),
			nil)
		return err
	})
}
