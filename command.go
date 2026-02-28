package oca

import (
	"errors"
	"fmt"
)

type Ocp1Command struct {
	CommandSize uint32
	Handle      uint32
	TargetONo   OcaONo
	MethodID    OcaMethodID
	Parameters  Ocp1Parameters
}

type Ocp1CommandData []Ocp1Command

func NewCommand(handle uint32, targetONo OcaONo, methodID OcaMethodID, parameters Ocp1Parameters) (Ocp1Command, error) {

	command := Ocp1Command{
		Handle:     handle,
		TargetONo:  targetONo,
		MethodID:   methodID,
		Parameters: parameters,
	}

	commandSize := uint32(17) // size of fixed fields
	parametersBytes, err := parameters.MarshalBinary()
	if err != nil {
		return Ocp1Command{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}
	commandSize += uint32(len(parametersBytes))
	command.CommandSize = commandSize

	return command, nil
}

func (c *Ocp1Command) UnmarshalBinary(data []byte) error {
	if len(data) < 17 {
		return errors.New("Ocp1Command: not enough data")
	}
	c.CommandSize = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	c.Handle = uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
	c.TargetONo = OcaONo(uint32(data[8])<<24 | uint32(data[9])<<16 | uint32(data[10])<<8 | uint32(data[11]))

	err := c.MethodID.UnmarshalBinary(data[12:16])
	if err != nil {
		return fmt.Errorf("Ocp1Command: failed to unmarshal MethodID: %w", err)
	}
	err = c.Parameters.UnmarshalBinary(data[16:c.CommandSize])
	if err != nil {
		return fmt.Errorf("Ocp1Command: failed to unmarshal Parameters: %w", err)
	}

	if c.Parameters.ParameterCount == 0 {
		return nil
	}
	c.Parameters.Parameters = make([]Ocp1Parameter, c.Parameters.ParameterCount)

	parameterDecoders, err := OcaObjectDecoders.GetParameterDecoders(c.TargetONo, c.MethodID.DefLevel, c.MethodID.MethodIndex)
	if err != nil {
		return fmt.Errorf("Ocp1Command: failed to get method parameter decoder: %w", err)
	}

	if len(parameterDecoders) != int(c.Parameters.ParameterCount) {
		return fmt.Errorf("Ocp1Command: expected %d parameter decoders got %d", c.Parameters.ParameterCount, len(parameterDecoders))
	}

	paramOffset := 0
	for i := 0; i < int(c.Parameters.ParameterCount); i++ {
		decoder := parameterDecoders[i]
		param, size, err := decoder(c.Parameters.Bytes[paramOffset:])
		if err != nil {
			return fmt.Errorf("Ocp1Command: failed to decode parameter %d: %w", i, err)
		}
		paramOffset += int(size)
		c.Parameters.Parameters[i] = param
	}
	return nil
}

func (c Ocp1Command) MarshalBinary() ([]byte, error) {
	bytes := make([]byte, 17)
	bytes[0] = byte(c.CommandSize >> 24)
	bytes[1] = byte((c.CommandSize >> 16) & 0xff)
	bytes[2] = byte((c.CommandSize >> 8) & 0xff)
	bytes[3] = byte(c.CommandSize & 0xff)

	bytes[4] = byte(c.Handle >> 24)
	bytes[5] = byte((c.Handle >> 16) & 0xff)
	bytes[6] = byte((c.Handle >> 8) & 0xff)
	bytes[7] = byte(c.Handle & 0xff)

	bytes[8] = byte(c.TargetONo >> 24)
	bytes[9] = byte((c.TargetONo >> 16) & 0xff)
	bytes[10] = byte((c.TargetONo >> 8) & 0xff)
	bytes[11] = byte(c.TargetONo & 0xff)

	methodIDBytes, err := c.MethodID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Ocp1Command: failed to marshal MethodID: %w", err)
	}
	bytes = append(bytes, methodIDBytes...)

	parametersBytes, err := c.Parameters.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("Ocp1Command: failed to marshal Parameters: %w", err)
	}
	bytes = append(bytes, parametersBytes...)
	return bytes, nil
}

func (d Ocp1CommandData) MarshalBinary() ([]byte, error) {
	var bytes []byte
	for _, cmd := range d {
		cmdBytes, err := cmd.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Ocp1CommandData: failed to marshal command: %w", err)
		}
		bytes = append(bytes, cmdBytes...)
	}
	return bytes, nil
}

func (c Ocp1Command) String() string {
	return fmt.Sprintf("{Size: %d, Handle: %d, TargetONo: %d, MethodID: %s, Parameters: %s}",
		c.CommandSize, c.Handle, c.TargetONo, c.MethodID.String(), c.Parameters.String())
}

func (c Ocp1Command) GetParameterDecoders() ([]ParameterDecoder, error) {
	if c.Parameters.ParameterCount == 0 {
		return nil, nil
	}
	parameterDecoders, err := OcaObjectDecoders.GetParameterDecoders(c.TargetONo, c.MethodID.DefLevel, c.MethodID.MethodIndex)
	if err != nil {
		return nil, fmt.Errorf("Ocp1Command: failed to get method parameter decoder: %w", err)
	}
	return parameterDecoders, nil
}
