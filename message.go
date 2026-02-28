package oca

import (
	"encoding"
	"errors"
	"fmt"
)

type Ocp1Data interface {
	encoding.BinaryMarshaler
}

type Ocp1MessagePdu struct {
	// Define the fields of the Ocp1MessagePdu struct here
	SyncVal uint8
	Header  Ocp1Header
	Data    Ocp1Data
}

func (m *Ocp1MessagePdu) UnmarshalBinary(data []byte) error {
	if len(data) < 10 {
		return errors.New("Ocp1MessagePdu: not enough data")
	}
	if data[0] != 0x3b {
		return errors.New("Ocp1MessagePdu: invalid start byte")
	}
	m.SyncVal = data[0]
	header := Ocp1Header{}
	err := header.UnmarshalBinary(data[1:10])
	if err != nil {
		return err
	}
	m.Header = header
	dataBytes := data[10:]

	switch m.Header.PduType {
	case Ocp1Cmd, Ocp1CmdRrq:
		commands := Ocp1CommandData{}
		commandOffset := 0
		for i := 0; i < int(m.Header.MessageCount); i++ {
			command := Ocp1Command{}
			err := command.UnmarshalBinary(dataBytes[commandOffset:])
			if err != nil {
				return fmt.Errorf("Ocp1MessagePdu: failed to unmarshal command %d: %w", i, err)
			}
			commands = append(commands, command)
			commandOffset += int(command.CommandSize)
		}
		m.Data = &commands
	case Ocp1KeepAlive:
		keepAliveData := Ocp1KeepAliveData{}
		err := keepAliveData.UnmarshalBinary(dataBytes)
		if err != nil {
			return err
		}
		m.Data = &keepAliveData
	default:
		return fmt.Errorf("Ocp1MessagePdu: unsupported PDU type: %d", m.Header.PduType)
	}
	return nil
}

func (m *Ocp1MessagePdu) MarshalBinary() ([]byte, error) {
	bytes := []byte{m.SyncVal}
	headerBytes, err := m.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, headerBytes...)
	dataBytes, err := m.Data.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, dataBytes...)
	return bytes, nil
}
