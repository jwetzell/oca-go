package oca

import (
	"errors"

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
