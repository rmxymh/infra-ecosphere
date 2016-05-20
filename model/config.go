package model

import (
	"os"
	"encoding/json"
	"log"
	"net"
)

type ConfigNode struct {
	BMCIP string
	VMName string
}

type ConfigBMCUser struct {
	Username string
	Password string
}

type Configuration struct {
	Nodes		[]ConfigNode
	BMCUsers	[]ConfigBMCUser
}

func LoadConfig(configFile string) {
	file, opError := os.Open(configFile)
	if opError != nil {
		log.Println("Config: Failed to open config file ", configFile, ", ignore...")
		return
	}

	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatalln("Config: Error: ", err)
	}

	// initialize BMCs and Instances
	for _, node := range configuration.Nodes {
		instance := AddInstnace(node.VMName)
		AddBMC(net.ParseIP(node.BMCIP), instance)
	}

	for _, user := range configuration.BMCUsers {
		log.Printf("Config: Add BMC User %s\n", user.Username)
		AddBMCUser(user.Username, user.Password)
	}
}