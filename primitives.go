package oca

import (
	"errors"
)

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
	data := make([]byte, 2+length)
	data[0] = byte(length >> 8)
	data[1] = byte(length & 0xff)
	data = append(data, b...)
	return data, nil
}

type OcaNotificationDeliveryMode uint8

func (m OcaNotificationDeliveryMode) MarshalBinary() ([]byte, error) {
	return []byte{byte(m)}, nil
}
