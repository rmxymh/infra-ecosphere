package main

import (
	"github.com/rmxymh/infra-ecosphere/ipmi"
	"github.com/rmxymh/infra-ecosphere/utils"
)

func main() {
	config := utils.LoadConfig("infra-ecosphere.cfg")
	EcospherePort = config.WebAPIPort
	ipmi.IPMI_CHASSIS_BOOT_OPTION_SetHandler(ipmi.BOOT_FLAG, SetBootDevice)
	ipmi.IPMIServerServiceRun()
}
