package main

import (
	"github.com/rmxymh/infra-ecosphere/ipmi"
	"github.com/rmxymh/infra-ecosphere/utils"
)

func main() {
	utils.LoadConfig("infra-ecosphere.cfg")
	ipmi.IPMIServerServiceRun()
}
