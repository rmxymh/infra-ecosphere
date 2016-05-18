package model

import (
	"net"
	"log"
)

type BMC struct {
	Addr net.IP
	PowerOn bool
}

var BMCs map[string]BMC

func init() {
	log.Println("Initialize BMC Map...")
	BMCs = make(map[string]BMC)
}

func AddBMC(ip net.IP) {
	newBMC := BMC{
		Addr: ip,
		PowerOn: false,
	}

	BMCs[ip.String()] = newBMC
	log.Println("Add new BMC with IP ", ip.String())
}

func RemoveBMC(ip net.IP) {
	_, ok := BMCs[ip.String()]

	if ok {
		delete(BMCs, ip.String())
	}
}

func GetBMC(ip net.IP) (BMC, bool) {
	obj, ok := BMCs[ip.String()]

	return obj, ok
}