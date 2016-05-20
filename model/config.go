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
		AddBMC(net.ParseIP(node.BMCIP))
		AddInstnace(node.VMName)

		instance, ok := GetInstance(node.VMName)
		if ok {
			BindInstance(net.ParseIP(node.BMCIP), instance)
			log.Printf("Config: Bind BMC %s onto Instance %s", node.BMCIP, instance.Name)
		} else {
			log.Fatalf("Config: Failed to bind BMC %s to Instance %s", node.BMCIP, node.VMName)
		}

	}

	for _, user := range configuration.BMCUsers {
		log.Printf("Config: Add BMC User %s\n", user.Username)
		AddBMCUser(user.Username, user.Password)
	}
}