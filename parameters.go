package oca

import (
	"encoding"
	"errors"
	"fmt"
)

type Ocp1Parameter interface {
	encoding.BinaryMarshaler
}

type Ocp1Parameters struct {
	ParameterCount uint8
	Bytes          []byte
	Parameters     []Ocp1Parameter
}

func (p *Ocp1Parameters) String() string {
	return fmt.Sprintf("%+v", p.Parameters)
}

func (p *Ocp1Parameters) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("Ocp1Parameters: not enough data")
	}
	p.ParameterCount = data[0]
	p.Bytes = data[1:]
	return nil
}

func (p Ocp1Parameters) MarshalBinary() ([]byte, error) {
	bytes := []byte{p.ParameterCount}
	for _, param := range p.Parameters {
		paramBytes, err := param.MarshalBinary()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, paramBytes...)
	}
	return bytes, nil
}

type ParameterDecoder func([]byte) (Ocp1Parameter, uint16, error)

type MethodDecoder map[uint16][]ParameterDecoder

type DefLevelDecoder map[uint16]MethodDecoder

type ObjectDecoder map[uint32]DefLevelDecoder

var ObjectDecoders = ObjectDecoder{
	4: {
		3: {
			8: []ParameterDecoder{
				func(data []byte) (Ocp1Parameter, uint16, error) {
					ocaEvent := OcaEvent{}
					err := ocaEvent.UnmarshalBinary(data)
					if err != nil {
						return nil, 0, err
					}
					return ocaEvent, 8, nil
				},
				func(data []byte) (Ocp1Parameter, uint16, error) {
					notificationDeliveryMode := OcaNotificationDeliveryMode(data[0])
					return notificationDeliveryMode, 1, nil
				},
				func(data []byte) (Ocp1Parameter, uint16, error) {
					ocaBlob := OcaBlob{}
					err := ocaBlob.UnmarshalBinary(data)
					if err != nil {
						return nil, 0, err
					}
					return ocaBlob, uint16(len(ocaBlob) + 1), nil
				},
			},
		},
	},
}
