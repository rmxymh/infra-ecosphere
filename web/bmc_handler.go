package web

import (
	"net/http"
	"encoding/json"
)

import (
	"github.com/rmxymh/infra-ecosphere/bmc"
)

type WebRespBMC struct {
	IP		string
	PowerStatus	string
}

func GetAllBMCs(writer http.ResponseWriter, request *http.Request) {
	RespBMCs := make([]WebRespBMC, 0)
	for _, b := range bmc.BMCs {
		status := "OFF"
		if b.IsPowerOn() {
			status = "ON"
		}
		RespBMCs = append(RespBMCs, WebRespBMC{
					IP: b.Addr.String(),
					PowerStatus: status,
		})
	}

	json.NewEncoder(writer).Encode(RespBMCs)
}

