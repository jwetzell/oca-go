package oca

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"
	"syscall"
	"time"
)

type Handle struct {
	TargetONo         OcaONo
	MethodID          OcaMethodID
	ParameterDecoders []ParameterDecoder
}

type OcaDevice struct {
	listener         *net.TCPListener
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	connections      []*net.TCPConn
	connectionsMu    sync.RWMutex
	quit             chan interface{}
	ocaMessageBuffer []byte
	handles          map[uint32][]Handle
	DeviceState      OcaDeviceState
}

func NewDevice(addr *net.TCPAddr) (*OcaDevice, error) {
	listener, err := net.ListenTCP("tcp", addr)

	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	return &OcaDevice{
		listener:         listener,
		handles:          make(map[uint32][]Handle),
		quit:             make(chan interface{}),
		connections:      []*net.TCPConn{},
		ocaMessageBuffer: []byte{},
	}, nil
}

func (d *OcaDevice) handleClient(client *net.TCPConn) {
	d.connectionsMu.Lock()
	d.connections = append(d.connections, client)
	d.connectionsMu.Unlock()
	defer client.Close()

	buffer := make([]byte, 1024)
ClientRead:
	for {
		select {
		case <-d.ctx.Done():
			client.Close()
			d.connectionsMu.Lock()
			for i := 0; i < len(d.connections); i++ {
				if d.connections[i] == client {
					d.connections = slices.Delete(d.connections, i, i+1)
					break
				}
			}
			d.connectionsMu.Unlock()
			return
		default:
			client.SetDeadline(time.Now().Add(time.Millisecond * 200))
			byteCount, err := client.Read(buffer)

			if err != nil {
				if opErr, ok := err.(*net.OpError); ok {
					//NOTE(jwetzell) we hit deadline
					if opErr.Timeout() {
						continue ClientRead
					}
					if errors.Is(opErr, syscall.ECONNRESET) {

						d.connectionsMu.Lock()
						for i := 0; i < len(d.connections); i++ {
							if d.connections[i] == client {
								d.connections = slices.Delete(d.connections, i, i+1)
								break
							}
						}
						d.connectionsMu.Unlock()
					}

				}

				if err.Error() == "EOF" {
					d.connectionsMu.Lock()
					for i := 0; i < len(d.connections); i++ {
						if d.connections[i] == client {
							d.connections = slices.Delete(d.connections, i, i+1)
							break
						}
					}
					d.connectionsMu.Unlock()
				}
				return
			}
			d.handleData(client, buffer[0:byteCount])
		}
	}
}

func (d *OcaDevice) handleOcaMessage(conn *net.TCPConn, message Ocp1MessagePdu) {
	fmt.Printf("Received OCA message: %+v\n", message)
	fmt.Printf("\tOCA Data: %+v\n", message.Data)
	switch message.Header.PduType {
	case Ocp1Cmd, Ocp1CmdRrq:
		commandData, ok := message.Data.(*Ocp1CommandData)
		if !ok {
			fmt.Printf("Failed to cast OCA data to Ocp1CommandData\n")
			break
		}
		for _, cmd := range *commandData {
			if message.Header.PduType == Ocp1CmdRrq {
				d.handleCommandRequiringResponse(conn, cmd)
			}
		}

	case Ocp1KeepAlive:
		_, ok := message.Data.(*Ocp1KeepAliveData)
		if !ok {
			fmt.Printf("Failed to cast OCA data to Ocp1KeepAliveData\n")
			break
		}
		keepAliveBytes, err := message.MarshalBinary()
		if err != nil {
			fmt.Printf("failed to marshal keep alive response: %s\n", err)
			break
		}
		_, err = conn.Write(keepAliveBytes)
		if err != nil {
			fmt.Printf("failed to write keep alive response: %s\n", err)
			break
		}
	case Ocp1Rsp:
		responseData, ok := message.Data.(*Ocp1ResponseData)
		if !ok {
			fmt.Printf("Failed to cast OCA data to Ocp1ResponseData\n")
			break
		}
		for _, rsp := range *responseData {
			fmt.Printf("\tResponse: %s\n", rsp.String())
		}
	default:
		fmt.Printf("Received OCA message: %+v\n", message)
		fmt.Printf("\tOCA Data: %+v\n", message.Data)
	}
}

func (d *OcaDevice) handleData(conn *net.TCPConn, data []byte) {
	fmt.Printf("handling data: %x\n", data)
	for _, dataByte := range data {
		if dataByte == 0x3b {
			if len(d.ocaMessageBuffer) > 0 {
				var pdu Ocp1MessagePdu
				err := pdu.UnmarshalBinary(d.ocaMessageBuffer)
				if err != nil {
					fmt.Printf("failed to unmarshal: %s\n", err)
					break
				}
				d.handleOcaMessage(conn, pdu)
				d.ocaMessageBuffer = []byte{}
			}
		}
		d.ocaMessageBuffer = append(d.ocaMessageBuffer, dataByte)
	}
}

func (d *OcaDevice) Start(ctx context.Context) error {

	if d.listener == nil {
		return errors.New("listener is nil")
	}

	d.ctx, d.cancel = context.WithCancel(ctx)

	d.wg.Add(1)

	go func() {
		<-d.ctx.Done()
		close(d.quit)
		d.listener.Close()
		fmt.Println("done")
	}()

	go func() {
	AcceptLoop:
		for {
			conn, err := d.listener.AcceptTCP()
			if err != nil {
				select {
				case <-d.quit:
					break AcceptLoop
				default:
					fmt.Printf("failed to accept connection: %s\n", err)
					continue
				}
			} else {
				d.wg.Go(func() {
					d.handleClient(conn)
					fmt.Printf("client donn: %s\n", conn.RemoteAddr().String())
				})
			}
		}
		d.wg.Done()
		d.wg.Wait()
	}()
	return nil
}

func (d *OcaDevice) handleCommandRequiringResponse(conn *net.TCPConn, cmd Ocp1Command) error {
	fmt.Printf("Received command requiring response: %+v\n", cmd)
	switch cmd.TargetONo {
	case OcaDeviceManager:
		switch cmd.MethodID.DefLevel {
		case 3:
			switch cmd.MethodID.MethodIndex {
			case 1:
				parameters, err := NewParameters([]Ocp1Parameter{OcaUint16(1)})
				if err != nil {
					return err
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return err
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return err
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return err
				}
				fmt.Printf("Writing response bytes to client at %s: %x\n", conn.RemoteAddr().String(), responseBytes)
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			case 2:
				parameters, err := NewParameters([]Ocp1Parameter{OcaModelGUID{
					Reserved:  OcaBlobFixed{0x01},
					MfrCode:   OcaBlobFixed{0x11, 0x22, 0x33},
					ModelCode: OcaBlobFixed{0x44, 0x55, 0x66, 0x77},
				}})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				enabledResponse, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				enabledResponseBytes, err := enabledResponse.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(enabledResponseBytes)
				if err != nil {
					return nil
				}
			case 3:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("abc123")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				conn.Write(responseBytes)
			case 4:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("oca-test-device")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			case 6:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("Test Manufacturer"), OcaString("Test Model"), OcaString("v0.0.0")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			case 7:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("TestDevice")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			case 9:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("TEST1234")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			case 11:
				parameters, err := NewParameters([]Ocp1Parameter{OcaBool(true)})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				enabledResponse, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				enabledResponseBytes, err := enabledResponse.MarshalBinary()
				if err != nil {
					return nil
				}
				conn.Write(enabledResponseBytes)
			case 13:
				parameters, err := NewParameters([]Ocp1Parameter{OcaDeviceState(0)})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				enabledResponse, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				enabledResponseBytes, err := enabledResponse.MarshalBinary()
				if err != nil {
					return nil
				}
				conn.Write(enabledResponseBytes)
			case 17:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("Test Message")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			case 20:
				parameters, err := NewParameters([]Ocp1Parameter{OcaString("ALPHA")})
				if err != nil {
					return nil
				}
				ocp1Response, err := NewResponse(cmd.Handle, 0, parameters)
				if err != nil {
					return nil
				}
				response, err := NewMessage(1, Ocp1Rsp, Ocp1ResponseData{
					ocp1Response,
				})
				if err != nil {
					return nil
				}
				responseBytes, err := response.MarshalBinary()
				if err != nil {
					return nil
				}
				_, err = conn.Write(responseBytes)
				if err != nil {
					return nil
				}
			default:
				return fmt.Errorf("unhandled method: %+v\n", cmd)
			}
		default:
			return fmt.Errorf("unhandled method deflevel: %+v\n", cmd)
		}
	default:
		return fmt.Errorf("unhandled targetono: %+v\n", cmd)
	}
	return nil
}

func (d *OcaDevice) HandleCommand(cmd Ocp1Command) {
}
