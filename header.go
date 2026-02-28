package oca

import "errors"

type Ocp1MessageType uint8

var (
	Ocp1Cmd       Ocp1MessageType = Ocp1MessageType(0x00)
	Ocp1CmdRrq    Ocp1MessageType = Ocp1MessageType(0x01)
	Ocp1Ntf1      Ocp1MessageType = Ocp1MessageType(0x02)
	Ocp1Rsp       Ocp1MessageType = Ocp1MessageType(0x03)
	Ocp1KeepAlive Ocp1MessageType = Ocp1MessageType(0x04)
	Ocp1Ntf2      Ocp1MessageType = Ocp1MessageType(0x05)
)

type Ocp1Header struct {
	ProtocolVersion uint16
	PduSize         uint32
	PduType         Ocp1MessageType
	MessageCount    uint16
}

func (h *Ocp1Header) UnmarshalBinary(data []byte) error {
	if len(data) < 9 {
		return errors.New("Ocp1Header: not enough data")
	}
	h.ProtocolVersion = uint16(data[0])<<8 | uint16(data[1])
	h.PduSize = uint32(data[2])<<24 | uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])
	pduType := Ocp1MessageType(data[6])
	switch pduType {
	case Ocp1Cmd, Ocp1CmdRrq, Ocp1Ntf1, Ocp1Rsp, Ocp1KeepAlive, Ocp1Ntf2:
		h.PduType = pduType
	default:
		return errors.New("Ocp1Header: invalid PDU type")
	}
	h.MessageCount = uint16(data[7])<<8 | uint16(data[8])
	return nil
}

func (h *Ocp1Header) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 9)
	bytes[0] = byte(h.ProtocolVersion >> 8)
	bytes[1] = byte(h.ProtocolVersion & 0xff)
	bytes[2] = byte(h.PduSize >> 24)
	bytes[3] = byte((h.PduSize >> 16) & 0xff)
	bytes[4] = byte((h.PduSize >> 8) & 0xff)
	bytes[5] = byte(h.PduSize & 0xff)
	bytes[6] = uint8(h.PduType)
	bytes[7] = byte(h.MessageCount >> 8)
	bytes[8] = byte(h.MessageCount & 0xff)
	return bytes, nil
}
