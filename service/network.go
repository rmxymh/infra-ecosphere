package service

import (
	"fmt"
	"os"
	"net"
	"bytes"
	"log"
)


import (
	"github.com/rmxymh/infra-ecosphere/protocol"
	"io"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}
}

func DeserializeAndExecute(buf io.Reader, addr *net.UDPAddr, server *net.UDPConn) {
	protocol.RMCPDeserializeAndExecute(buf, addr, server)
}

func NetworkServiceRun() {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.1.1:623")
	CheckError(err)

	server, err := net.ListenUDP("udp", serverAddr)
	CheckError(err)
	defer server.Close()

	buf := make([]byte, 1024)
	for {
		_, addr, _ := server.ReadFromUDP(buf)
		log.Println("Receive a UDP packet from ", addr.IP.String(), ":", addr.Port)

		bytebuf := bytes.NewBuffer(buf)
		DeserializeAndExecute(bytebuf, addr, server)
	}
}