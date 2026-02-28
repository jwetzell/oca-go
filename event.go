package oca

import "errors"

type OcaEvent struct {
	EmitterONo OcaONo
	EventID    OcaEventID
}

func (e *OcaEvent) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("OcaEvent: not enough data")
	}
	e.EmitterONo = OcaONo(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))

	err := e.EventID.UnmarshalBinary(data[4:8])
	if err != nil {
		return err
	}
	return nil
}

func (e *OcaEvent) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 8)
	bytes[0] = byte(e.EmitterONo >> 24)
	bytes[1] = byte((e.EmitterONo >> 16) & 0xff)
	bytes[2] = byte((e.EmitterONo >> 8) & 0xff)
	bytes[3] = byte(e.EmitterONo & 0xff)

	eventIDBytes, err := e.EventID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, eventIDBytes...)
	return bytes, nil
}

type OcaEventID struct {
	DefLevel   uint16
	EventIndex uint16
}

func (e *OcaEventID) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("OcaEventID: not enough data")
	}
	e.DefLevel = uint16(data[0])<<8 | uint16(data[1])
	e.EventIndex = uint16(data[2])<<8 | uint16(data[3])
	return nil
}

func (e *OcaEventID) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 4)
	bytes[0] = byte(e.DefLevel >> 8)
	bytes[1] = byte(e.DefLevel & 0xff)
	bytes[2] = byte(e.EventIndex >> 8)
	bytes[3] = byte(e.EventIndex & 0xff)
	return bytes, nil
}
