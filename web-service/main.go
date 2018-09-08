package main

import (
	"fmt"
	"os"

	"github.com/akkagao/citizens/web-service/blockchain"
)

func main() {
	// Definition of the Fabric SDK properties
	fSetup := blockchain.FabricSetup{
		// Network parameters
		OrdererID: "orderer.citizens.com",

		// Channel parameters
		ChannelID:     "citizens",
		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/akkagao/citizens/fabric-service/artifacts/citizens.tx",

		// Chaincode parameters
		ChainCodeID:      "citizens-service",
		ChainCodeVersion: "0",
		ChaincodeGoPath:  os.Getenv("GOPATH"),
		ChaincodePath:    "github.com/akkagao/citizens/chaincode/",
		OrgAdmin:         "Admin",
		OrgName:          "FBI",
		ConfigFile:       "config.yaml",

		// User parameters
		UserName: "User1",
	}

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
		return
	}
	// Close SDK
	defer fSetup.CloseSDK()

	// Install and instantiate the chaincode
	err = fSetup.InstallAndInstantiateCC()
	if err != nil {
		fmt.Printf("Unable to install and instantiate the chaincode: %v\n", err)
		return
	}

	// Launch the web application listening
	// app := &controllers.Application{
	// 	Fabric: &fSetup,
	// }
	// web.Serve(app)
	fSetup.RegisterUser()
	fSetup.QueryUser()

}
