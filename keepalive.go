package oca

import (
	"errors"
	"fmt"
)

type Ocp1KeepAliveData struct {
	HeartbeatTimeout uint32
	Seconds          bool
}

func (d Ocp1KeepAliveData) String() string {
	if d.Seconds {
		return fmt.Sprintf("{HeartbeatTimeout: %d seconds}", d.HeartbeatTimeout)
	}
	return fmt.Sprintf("{HeartbeatTimeout: %d milliseconds}", d.HeartbeatTimeout)
}

func (d *Ocp1KeepAliveData) UnmarshalBinary(data []byte) error {
	if len(data) == 2 {
		heartbeatTimeout := uint32(data[0])<<8 | uint32(data[1])
		d.HeartbeatTimeout = heartbeatTimeout
		d.Seconds = true
	} else if len(data) == 4 {
		heartbeatTimeout := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		d.HeartbeatTimeout = heartbeatTimeout
		d.Seconds = false
	} else {
		return errors.New("Ocp1KeepAliveData: invalid data length")
	}
	return nil
}

func (d Ocp1KeepAliveData) MarshalBinary() ([]byte, error) {
	if d.Seconds {
		bytes := make([]byte, 2)
		bytes[0] = byte(d.HeartbeatTimeout >> 8)
		bytes[1] = byte(d.HeartbeatTimeout & 0xff)
		return bytes, nil
	}

	bytes := make([]byte, 4)
	bytes[0] = byte(d.HeartbeatTimeout >> 24)
	bytes[1] = byte((d.HeartbeatTimeout >> 16) & 0xff)
	bytes[2] = byte((d.HeartbeatTimeout >> 8) & 0xff)
	bytes[3] = byte(d.HeartbeatTimeout & 0xff)
	return bytes, nil
}
