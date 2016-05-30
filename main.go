package main

import (
	"github.com/rmxymh/infra-ecosphere/ipmi"
	"github.com/rmxymh/infra-ecosphere/utils"
	"github.com/rmxymh/infra-ecosphere/web"
)

func main() {
	utils.LoadConfig("infra-ecosphere.cfg")
	go ipmi.IPMIServerServiceRun()
	web.WebAPIServiceRun()
}
