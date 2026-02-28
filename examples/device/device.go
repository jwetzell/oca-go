package main

import (
	"context"
	"net"
	"os"
	"os/signal"

	"github.com/jwetzell/oca-go"
)

func main() {
	device, err := oca.NewDevice(&net.TCPAddr{Port: 65000})
	if err != nil {
		panic(err)
	}

	signalContext, signalCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	device.Start(signalContext)
	<-signalContext.Done()
	signalCancel()
}
