package oca

import "errors"

type Ocp1Parameters struct {
	ParameterCount uint8
	Bytes          []byte
}

func (p *Ocp1Parameters) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("Ocp1Parameters: not enough data")
	}
	p.ParameterCount = data[0]
	p.Bytes = data[1:]
	return nil
}

func (p *Ocp1Parameters) MarshalBinary() ([]byte, error) {
	bytes := []byte{p.ParameterCount}
	// bytes = append(bytes, p.Parameters...)
	return bytes, nil
}

type ParameterDecoder func([]byte) (any, uint16, error)

type MethodDecoder map[uint16][]ParameterDecoder

type DefLevelDecoder map[uint16]MethodDecoder

type ObjectDecoder map[uint32]DefLevelDecoder

var ObjectDecoders = ObjectDecoder{
	4: {
		3: {
			1: []ParameterDecoder{
				func(data []byte) (any, uint16, error) {
					ocaEventId := OcaEventID{}
					err := ocaEventId.UnmarshalBinary(data)
					if err != nil {
						return nil, 0, err
					}
					return ocaEventId, 4, nil
				},
				func(data []byte) (any, uint16, error) {
					ocaMethod := OcaMethod{}
					err := ocaMethod.UnmarshalBinary(data)
					if err != nil {
						return nil, 0, err
					}
					return ocaMethod, 8, nil
				},
				func(data []byte) (any, uint16, error) {
					ocaBlob := OcaBlob{}
					err := ocaBlob.UnmarshalBinary(data)
					if err != nil {
						return nil, 0, err
					}
					return ocaBlob, uint16(len(ocaBlob) + 1), nil
				},
				func(data []byte) (any, uint16, error) {
					notificationDeliveryMode := OcaNotificationDeliveryMode(data[0])
					return notificationDeliveryMode, 1, nil
				},
				func(data []byte) (any, uint16, error) {
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
