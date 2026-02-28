package main

import (
	"fmt"
	"net"

	"github.com/jwetzell/oca-go"
)

func main() {
	listenTCP("0.0.0.0:5555")
}

func handleTCPConnection(conn net.Conn) {

	// go handleSLIP(slip, format)

	defer conn.Close()
	buffer := make([]byte, 1024)

	ocaMessageBytes := []byte{}

	for {
		bytesRead, err := conn.Read(buffer)

		if err != nil {
			fmt.Println(err)
			return
		}

		for i := 0; i < bytesRead; i++ {
			if buffer[i] == 0x3b {
				if len(ocaMessageBytes) > 0 {
					handleOCAMessage(conn, ocaMessageBytes)
					ocaMessageBytes = []byte{}
				}
			}
			ocaMessageBytes = append(ocaMessageBytes, buffer[i])
		}
	}
}

func handleOCAMessage(conn net.Conn, data []byte) {
	var pdu oca.Ocp1MessagePdu
	err := pdu.UnmarshalBinary(data)
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
		}
	case oca.Ocp1KeepAlive:
		keepAliveData, ok := pdu.Data.(*oca.Ocp1KeepAliveData)
		if !ok {
			fmt.Printf("Failed to cast OCA data to Ocp1KeepAliveData\n")
			return
		}
		fmt.Printf("Received KeepAlive message: %s\n", keepAliveData.String())
		// echo back
		conn.Write(data)
	default:
		fmt.Printf("Received OCA message: %+v\n", pdu)
		fmt.Printf("\tOCA Data: %+v\n", pdu.Data)
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
