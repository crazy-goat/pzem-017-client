package main

import (
	"encoding/binary"
	"fmt"
	"github.com/goburrow/modbus"
)

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