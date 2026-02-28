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

func NewParameters(parameters []Ocp1Parameter) (Ocp1Parameters, error) {
	if len(parameters) > 255 {
		return Ocp1Parameters{}, fmt.Errorf("too many parameters: %d", len(parameters))
	}
	return Ocp1Parameters{
		ParameterCount: uint8(len(parameters)),
		Parameters:     parameters,
	}, nil
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

type ParameterDecoders map[uint16][]ParameterDecoder

type DefLevelDecoders map[uint16]ParameterDecoders

type ObjectDecoders map[OcaONo]DefLevelDecoders

var OcaObjectDecoders = ObjectDecoders{
	OcaDeviceManager: {
		3: {
			12: []ParameterDecoder{ // SetEnabled(OcaBool)
				func(data []byte) (Ocp1Parameter, uint16, error) {
					ocaBool := OcaBool(true)
					err := ocaBool.UnmarshalBinary(data)
					if err != nil {
						return nil, 0, err
					}
					return ocaBool, 1, nil
				},
			},
		},
	},
	OcaSubscriptionManager: {
		3: {
			8: []ParameterDecoder{ // AddSubscription2(OcaEvent, OcaNotificationDeliveryMode, OcaBlob)
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

func (od ObjectDecoders) GetParameterDecoders(targetONo OcaONo, defLevel uint16, methodIndex uint16) ([]ParameterDecoder, error) {
	objectDecoder, ok := od[targetONo]
	if !ok {
		return nil, fmt.Errorf("Ocp1Command: no decoder found for TargetONo %d.%d.%d", targetONo, defLevel, methodIndex)
	}

	defLevelDecoder, ok := objectDecoder[defLevel]
	if !ok {
		return nil, fmt.Errorf("Ocp1Command: no decoder found for DefLevel %d.%d.%d", targetONo, defLevel, methodIndex)
	}

	parameterDecoders, ok := defLevelDecoder[methodIndex]
	if !ok {
		return nil, fmt.Errorf("Ocp1Command: no decoder found for MethodIndex %d.%d.%d", targetONo, defLevel, methodIndex)
	}

	return parameterDecoders, nil
}
