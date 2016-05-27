package main

import (
	"github.com/rmxymh/infra-ecosphere/ipmi"
	"github.com/rmxymh/infra-ecosphere/utils"
)

func main() {
	utils.LoadConfig("infra-ecosphere.cfg")
	ipmi.IPMI_CHASSIS_BOOT_OPTION_SetHandler(ipmi.BOOT_FLAG, SetBootDevice)
	ipmi.IPMIServerServiceRun()
}
