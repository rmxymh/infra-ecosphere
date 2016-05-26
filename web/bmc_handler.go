package web

import (
	"net/http"
	"encoding/json"
)

import (
	"github.com/rmxymh/infra-ecosphere/bmc"
	"github.com/gorilla/mux"
	"strings"
	"fmt"
	"net"
)

type WebRespBMC struct {
	IP		string
	PowerStatus	string
}

type WebRespBMCList struct {
	BMCs	[]WebRespBMC
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

	resp := WebRespBMCList{
		BMCs: RespBMCs,
	}
	json.NewEncoder(writer).Encode(resp)
}

type WebReqPowerOp struct {
	Operation	string
}

type WebRespPowerOp struct {
	IP		string
	Operation	string
	Status		string
}

func SetPowerStatus(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	resp := WebRespPowerOp{}
	resp.IP = vars["bmcip"]

	bmcobj, ok := bmc.GetBMC(net.ParseIP(resp.IP))
	if ! ok {
		resp.Status = fmt.Sprintf("BMC %s does not exist.", resp.IP)
	} else {
		powerOpReq := WebReqPowerOp{}
		err := json.NewDecoder(request.Body).Decode(&powerOpReq)

		if err != nil {
			resp.Operation = "Unknown"
			resp.Status = err.Error()
		} else {
			resp.Operation = strings.ToUpper(powerOpReq.Operation)
			switch resp.Operation {
			case "ON":
				bmcobj.PowerOn()
			case "OFF":
				bmcobj.PowerOff()
			case "SOFT":
				bmcobj.PowerSoft()
			case "RESET":
				bmcobj.PowerReset()
			case "CYCLE":
				bmcobj.PowerReset()
			}
			resp.Status = "OK"
		}
	}

	json.NewEncoder(writer).Encode(resp)
}