package oca

import (
	"errors"
	"fmt"
)

type OcaBool bool

func (b *OcaBool) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("OcaBool: not enough data")
	}
	*b = data[0] != 0
	return nil
}

func (b OcaBool) MarshalBinary() ([]byte, error) {
	var value byte
	if b {
		value = 1
	} else {
		value = 0
	}
	return []byte{value}, nil
}

type OcaStatus uint8

type OcaONo uint32

type OcaBlob []byte

func (b *OcaBlob) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("OcaBlob: not enough data")
	}
	length := uint16(data[0])<<8 | uint16(data[1])
	if len(data) < int(2+length) {
		return errors.New("OcaBlob: not enough data for blob content")
	}
	*b = data[2 : 2+length]
	return nil
}

func (b OcaBlob) MarshalBinary() ([]byte, error) {
	length := len(b)
	if length > 65535 {
		return nil, errors.New("OcaBlob: blob too large to marshal")
	}
	data := []byte{byte(length >> 8), byte(length & 0xff)}
	data = append(data, b...)

	return data, nil
}

type OcaNotificationDeliveryMode uint8

func (m OcaNotificationDeliveryMode) MarshalBinary() ([]byte, error) {
	return []byte{byte(m)}, nil
}

type OcaString string

func (s *OcaString) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("OcaString: not enough data")
	}
	length := uint16(data[0])<<8 | uint16(data[1])
	if len(data) < int(2+length) {
		return errors.New("OcaString: not enough data for string content")
	}
	*s = OcaString(data[2 : 2+length])
	return nil
}

func (s OcaString) MarshalBinary() ([]byte, error) {
	length := len(s)
	if length > 65535 {
		return nil, errors.New("OcaString: string too large to marshal")
	}
	data := []byte{byte(length >> 8), byte(length & 0xff)}
	data = append(data, []byte(s)...)

	return data, nil
}

type OcaDeviceState uint16

func (s OcaDeviceState) MarshalBinary() ([]byte, error) {
	return []byte{byte(s >> 8), byte(s & 0xff)}, nil
}

type OcaModelGUID struct {
	Reserved  OcaBlobFixed
	MfrCode   OcaBlobFixed
	ModelCode OcaBlobFixed
}

func (g *OcaModelGUID) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("OcaModelGUID: not enough data")
	}
	g.Reserved = OcaBlobFixed(data[0:1])
	g.MfrCode = OcaBlobFixed(data[1 : 1+3])
	g.ModelCode = OcaBlobFixed(data[4:8])
	return nil
}

func (g OcaModelGUID) MarshalBinary() ([]byte, error) {
	reservedBytes, err := g.Reserved.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("OcaModelGUID: failed to marshal Reserved: %w", err)
	}
	mfrCodeBytes, err := g.MfrCode.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("OcaModelGUID: failed to marshal MfrCode: %w", err)
	}
	modelCodeBytes, err := g.ModelCode.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("OcaModelGUID: failed to marshal ModelCode: %w", err)
	}

	data := append(reservedBytes, mfrCodeBytes...)
	data = append(data, modelCodeBytes...)
	return data, nil
}

type OcaBlobFixed []byte

func (b *OcaBlobFixed) UnmarshalBinary(data []byte) error {
	*b = data
	return nil
}

func (b OcaBlobFixed) MarshalBinary() ([]byte, error) {
	return []byte(b), nil
}

type OcaUint16 uint16

func (u OcaUint16) MarshalBinary() ([]byte, error) {
	return []byte{byte(u >> 8), byte(u & 0xff)}, nil
}
