package oca

import (
	"errors"
	"fmt"
	"net"
)

type OcaBase struct {
	Ono  OcaONo
	conn net.Conn
}

func (o *OcaBase) SendCommandRequiringResponse(methodID OcaMethodID, params Ocp1Parameters) error {
	if o.conn == nil {
		return errors.New("ObjectBase: connection not set")
	}

	commandData, err := NewCommand(1, o.Ono, methodID, params)
	if err != nil {
		return fmt.Errorf("ObjectBase: failed to create command: %w", err)
	}

	message, err := NewMessage(1, Ocp1CmdRrq, commandData)
	if err != nil {
		return fmt.Errorf("ObjectBase: failed to create message: %w", err)
	}
	messageBytes, err := message.MarshalBinary()
	if err != nil {
		return fmt.Errorf("ObjectBase: failed to marshal message: %w", err)
	}

	_, err = o.conn.Write(messageBytes)
	if err != nil {
		return fmt.Errorf("ObjectBase: failed to send message: %w", err)
	}

	return nil
}

func (o *OcaBase) SendResponse(handle uint32, status OcaStatus, params Ocp1Parameters) error {
	if o.conn == nil {
		return errors.New("ObjectBase: connection not set")
	}

	response, err := NewResponse(handle, status, params)

	if err != nil {
		return fmt.Errorf("ObjectBase: failed to create response: %w", err)
	}

	responseData := Ocp1ResponseData{response}

	message, err := NewMessage(1, Ocp1Rsp, responseData)

	if err != nil {
		return fmt.Errorf("ObjectBase: failed to create message: %w", err)
	}
	messageBytes, err := message.MarshalBinary()
	if err != nil {
		return fmt.Errorf("ObjectBase: failed to marshal message: %w", err)
	}

	_, err = o.conn.Write(messageBytes)
	if err != nil {
		return fmt.Errorf("ObjectBase: failed to send message: %w", err)
	}

	return nil

}
