package oca

import (
	"errors"
	"fmt"
)

type Ocp1Response struct {
	ResponseSize uint32
	Handle       uint32
	StatusCode   OcaStatus
	Parameters   Ocp1Parameters
}

type Ocp1ResponseData []Ocp1Response

func NewResponse(handle uint32, status OcaStatus, parameters Ocp1Parameters) (Ocp1Response, error) {

	response := Ocp1Response{
		Handle:     handle,
		StatusCode: status,
		Parameters: parameters,
	}

	responseSize := uint32(9) // size of fixed fields
	parametersBytes, err := parameters.MarshalBinary()
	if err != nil {
		return Ocp1Response{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}
	responseSize += uint32(len(parametersBytes))
	response.ResponseSize = responseSize

	return response, nil
}

func (r *Ocp1Response) UnmarshalBinary(data []byte) error {
	if len(data) < 9 {
		return errors.New("Ocp1Response: not enough data")
	}
	r.ResponseSize = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	r.Handle = uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
	r.StatusCode = OcaStatus(data[8])

	err := r.Parameters.UnmarshalBinary(data[9:r.ResponseSize])
	if err != nil {
		return fmt.Errorf("Ocp1Response: failed to unmarshal Parameters: %w", err)
	}
	// TODO(jwetzell): unmarshal parameters

	return nil
}

func (r Ocp1Response) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 9)
	bytes[0] = byte(r.ResponseSize >> 24)
	bytes[1] = byte((r.ResponseSize >> 16) & 0xff)
	bytes[2] = byte((r.ResponseSize >> 8) & 0xff)
	bytes[3] = byte(r.ResponseSize & 0xff)

	bytes[4] = byte(r.Handle >> 24)
	bytes[5] = byte((r.Handle >> 16) & 0xff)
	bytes[6] = byte((r.Handle >> 8) & 0xff)
	bytes[7] = byte(r.Handle & 0xff)

	bytes[8] = byte(r.StatusCode)

	parametersBytes, err := r.Parameters.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Ocp1Response: failed to marshal Parameters: %w", err)
	}
	bytes = append(bytes, parametersBytes...)

	return bytes, nil
}

func (d Ocp1ResponseData) MarshalBinary() ([]byte, error) {
	var bytes []byte
	for _, cmd := range d {
		cmdBytes, err := cmd.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Ocp1ResponseData: failed to marshal command: %w", err)
		}
		bytes = append(bytes, cmdBytes...)
	}
	return bytes, nil
}

func (r *Ocp1Response) String() string {
	return fmt.Sprintf("{Size: %d, Handle: %d, Status: %d, Parameters: %s}",
		r.ResponseSize, r.Handle, r.StatusCode, r.Parameters.String())
}
