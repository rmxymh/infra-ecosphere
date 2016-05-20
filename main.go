package main

import (
	"github.com/rmxymh/infra-ecosphere/service"
	"github.com/rmxymh/infra-ecosphere/model"
)

func main() {
	// make default data
	//model.AddBMCUser("admin", "admin")
	//model.AddBMC(net.ParseIP("127.0.1.1"))
	//model.AddBMC(net.ParseIP("127.0.1.2"))
	model.LoadConfig("infra-ecosphere.cfg")

	service.IPMIServerServiceRun()
}
