package main

import (
	"github.com/rmxymh/infra-ecosphere/service"
	"github.com/rmxymh/infra-ecosphere/model"
)

func main() {
	// make default data
	model.AddBMCUser("admin", "admin")

	service.NetworkServiceRun()
}
