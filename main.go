package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

import (
	"github.com/rmxymh/infra-ecosphere/protocol"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}
}

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.1.1:623")
	CheckError(err)

	server, err := net.ListenUDP("udp", serverAddr)
	CheckError(err)
	defer server.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := server.ReadFromUDP(buf)
		rmcp := protocol.RemoteManagementControlProtocol{}
		asf := protocol.AlertStandardFormat{}

		bytebuf := bytes.NewBuffer(buf)
		binary.Read(bytebuf, binary.BigEndian, &rmcp)
		if rmcp.Class == protocol.RMCP_CLASS_ASF {
			binary.Read(bytebuf, binary.BigEndian, &asf)
		}

		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
		fmt.Println(rmcp)
		fmt.Println(asf)
	}
}
