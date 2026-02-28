package main

import (
	"fmt"
	"net"

	"github.com/jwetzell/oca-go"
)

func main() {
	listenTCP("0.0.0.0:5555")
}

type Handle struct {
	TargetONo         oca.OcaONo
	MethodID          oca.OcaMethodID
	ParameterDecoders []oca.ParameterDecoder
}

func handleTCPConnection(conn net.Conn) {

	// go handleSLIP(slip, format)

	defer conn.Close()
	buffer := make([]byte, 1024)

	ocaMessageBytes := []byte{}

	handles := map[uint32][]Handle{}

	for {
		bytesRead, err := conn.Read(buffer)

		if err != nil {
			fmt.Println(err)
			return
		}

		for i := 0; i < bytesRead; i++ {
			if buffer[i] == 0x3b {
				if len(ocaMessageBytes) > 0 {
					var pdu oca.Ocp1MessagePdu
					err := pdu.UnmarshalBinary(ocaMessageBytes)
					if err != nil {
						fmt.Printf("failed to unmarshal: %s\n", err)
						return
					}
					switch pdu.Header.PduType {
					case oca.Ocp1Cmd, oca.Ocp1CmdRrq:
						fmt.Println("Received Command message:")
						commandData, ok := pdu.Data.(*oca.Ocp1CommandData)
						if !ok {
							fmt.Printf("Failed to cast OCA data to Ocp1CommandData\n")
							return
						}
						for _, cmd := range *commandData {
							fmt.Printf("\tCommand: %s\n", cmd.String())
							if pdu.Header.PduType == oca.Ocp1CmdRrq {
								parameterDecoders, err := cmd.GetParameterDecoders()
								if err != nil {
									fmt.Printf("Failed to get parameter decoders for command: %s\n", err)
									break
								}
								handles[cmd.Handle] = append(handles[cmd.Handle], Handle{
									TargetONo:         cmd.TargetONo,
									MethodID:          cmd.MethodID,
									ParameterDecoders: parameterDecoders,
								})
								fmt.Printf("\tRegistered handle %d with %d parameter decoders\n", cmd.Handle, len(parameterDecoders))
							}
						}

					case oca.Ocp1KeepAlive:
						keepAliveData, ok := pdu.Data.(*oca.Ocp1KeepAliveData)
						if !ok {
							fmt.Printf("Failed to cast OCA data to Ocp1KeepAliveData\n")
							return
						}
						fmt.Printf("Received KeepAlive message: %s\n", keepAliveData.String())
						// echo back
						conn.Write(ocaMessageBytes)
					default:
						fmt.Printf("Received OCA message: %+v\n", pdu)
						fmt.Printf("\tOCA Data: %+v\n", pdu.Data)
					}
					ocaMessageBytes = []byte{}
				}
			}
			ocaMessageBytes = append(ocaMessageBytes, buffer[i])
		}
	}
}

func listenTCP(netAddress string) {
	socket, err := net.Listen("tcp4", netAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer socket.Close()

	for {
		conn, err := socket.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleTCPConnection(conn)
	}
}
