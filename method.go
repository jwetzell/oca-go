package oca

import "errors"

type OcaMethod struct {
	ONo      OcaONo
	MethodID OcaMethodID
}

func (m *OcaMethod) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("OcaMethod: not enough data")
	}
	m.ONo = OcaONo(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
	err := m.MethodID.UnmarshalBinary(data[4:8])
	if err != nil {
		return err
	}
	return nil
}

func (m *OcaMethod) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 8)
	bytes[0] = byte(m.ONo >> 24)
	bytes[1] = byte((m.ONo >> 16) & 0xff)
	bytes[2] = byte((m.ONo >> 8) & 0xff)
	bytes[3] = byte(m.ONo & 0xff)

	methodIDBytes, err := m.MethodID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, methodIDBytes...)
	return bytes, nil
}

type OcaMethodID struct {
	DefLevel    uint16
	MethodIndex uint16
}

func (c *OcaMethodID) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("OcaMethodID: not enough data")
	}
	c.DefLevel = uint16(data[0])<<8 | uint16(data[1])
	c.MethodIndex = uint16(data[2])<<8 | uint16(data[3])
	return nil
}

func (c *OcaMethodID) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 4)
	bytes[0] = byte(c.DefLevel >> 8)
	bytes[1] = byte(c.DefLevel & 0xff)
	bytes[2] = byte(c.MethodIndex >> 8)
	bytes[3] = byte(c.MethodIndex & 0xff)
	return bytes, nil
}
