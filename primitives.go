package oca

import "errors"

type OcaBlob []byte

func (b *OcaBlob) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("OcaBlob: not enough data")
	}
	length := data[0]
	if len(data) < int(1+length) {
		return errors.New("OcaBlob: not enough data for blob content")
	}
	*b = data[1 : 1+length]
	return nil
}

func (b *OcaBlob) MarshalBinary() ([]byte, error) {
	length := len(*b)
	if length > 255 {
		return nil, errors.New("OcaBlob: blob too large")
	}
	data := make([]byte, 1+length)
	data[0] = byte(length)
	data = append(data, *b...)
	return data, nil
}
