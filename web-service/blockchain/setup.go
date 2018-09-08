package blockchain

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"
)

// FabricSetup implementation
type FabricSetup struct {
	ConfigFile       string
	OrgID            string
	OrdererID        string
	ChannelID        string
	ChainCodeID      string
	ChainCodeVersion string
	initialized      bool
	ChannelConfig    string
	ChaincodeGoPath  string
	ChaincodePath    string
	OrgAdmin         string
	OrgName          string
	UserName         string
	client           *channel.Client
	admin            *resmgmt.Client
	sdk              *fabsdk.FabricSDK
	event            *event.Client
}

// Initialize reads the configuration file and sets up the client, chain and event hub
func (setup *FabricSetup) Initialize() error {

	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("sdk already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		return errors.WithMessage(err, "failed to create SDK")
	}
	setup.sdk = sdk
	fmt.Println("SDK created")

	// The resource management client is responsible for managing channels (create/update channel)
	resourceManagerClientContext := setup.sdk.Context(fabsdk.WithUser(setup.OrgAdmin), fabsdk.WithOrg(setup.OrgName))
	if err != nil {
		return errors.WithMessage(err, "failed to load Admin identity")
	}
	resMgmtClient, err := resmgmt.New(resourceManagerClientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel management client from Admin identity")
	}
	setup.admin = resMgmtClient
	fmt.Println("Ressource management client created")

	// The MSP client allow us to retrieve user information from their identity, like its signing identity which we will need to save the channel
	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(setup.OrgName))
	if err != nil {
		return errors.WithMessage(err, "failed to create MSP client")
	}
	adminIdentity, err := mspClient.GetSigningIdentity(setup.OrgAdmin)
	if err != nil {
		return errors.WithMessage(err, "failed to get admin signing identity")
	}

	channelHasInstall := false
	// 查询已经存在的channel
	channelRes, err := setup.admin.QueryChannels(resmgmt.WithTargetEndpoints("peer0.fbi.citizens.com"))
	if err != nil {
		return errors.WithMessage(err, "failed to Query channel")
	}

	if channelRes != nil {
		for _, channel := range channelRes.Channels {
			if strings.EqualFold(setup.ChannelID, channel.ChannelId) {
				channelHasInstall = true
			}
		}
	}

	fmt.Println("channelHasInstall:", channelHasInstall)

	// 如果没有同名的channel 再安装
	if !channelHasInstall {
		req := resmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfigPath: setup.ChannelConfig, SigningIdentities: []msp.SigningIdentity{adminIdentity}}
		// 创建channel
		txID, err := setup.admin.SaveChannel(req, resmgmt.WithOrdererEndpoint(setup.OrdererID))
		if err != nil || txID.TransactionID == "" {
			return errors.WithMessage(err, "failed to save channel")
		}
		fmt.Println("Channel created")

		// 加入channel
		// Make admin user join the previously created channel
		if err = setup.admin.JoinChannel(setup.ChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(setup.OrdererID)); err != nil {
			return errors.WithMessage(err, "failed to make admin join channel")
		}
		fmt.Println("Channel joined")
	} else {
		fmt.Println("Channel already exist")
	}

	fmt.Println("Initialization Successful")
	setup.initialized = true
	return nil
}

func (setup *FabricSetup) InstallAndInstantiateCC() error {

	// Create the chaincode package that will be sent to the peers
	ccPkg, err := packager.NewCCPackage(setup.ChaincodePath, setup.ChaincodeGoPath)
	if err != nil {
		return errors.WithMessage(err, "failed to create chaincode package")
	}
	fmt.Println("ccPkg created")

	ccHasInstall := false
	// 查询已经安装的CC
	ccInstalledRes, err := setup.admin.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints("peer0.fbi.citizens.com"))
	if err != nil {
		return errors.WithMessage(err, "failed to Query Installed chaincode")
	}

	if ccInstalledRes != nil {
		for _, cc := range ccInstalledRes.Chaincodes {
			if strings.EqualFold(cc.Name, setup.ChainCodeID) {
				ccHasInstall = true
			}
		}
	}

	fmt.Println("ccHasInstall", ccHasInstall)

	if !ccHasInstall {
		// 安装链码（智能合约）
		// Install example cc to org peers
		installCCReq := resmgmt.InstallCCRequest{Name: setup.ChainCodeID, Path: setup.ChaincodePath, Version: setup.ChainCodeVersion, Package: ccPkg}
		_, err = setup.admin.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			return errors.WithMessage(err, "failed to install chaincode")
		}
		fmt.Println("Chaincode install success")
	} else {
		fmt.Println("Chaincode already exist")
	}

	ccHasInstantiate := false
	// 查询已经实例化的链码
	// ccInstantiatedRes, err := setup.admin.QueryInstantiatedChaincodes(setup.ChannelID, resmgmt.WithOrdererEndpoint(setup.OrdererID))
	ccInstantiatedRes, err := setup.admin.QueryInstantiatedChaincodes(setup.ChannelID, resmgmt.WithTargetEndpoints("peer0.fbi.citizens.com"))

	if ccInstantiatedRes.Chaincodes != nil && len(ccInstantiatedRes.Chaincodes) > 0 {
		for _, chaincodeInfo := range ccInstantiatedRes.Chaincodes {
			fmt.Println(chaincodeInfo)
			if strings.EqualFold(chaincodeInfo.Name, setup.ChainCodeID) {
				ccHasInstantiate = true
			}
		}
	}

	// could not get chConfig cache reference:read configuration for channel peers failed

	// Set up chaincode policy
	// ccPolicy := cauthdsl.SignedByAnyMember([]string{"fbi.citizens.com"})
	if !ccHasInstantiate {
		// 这里的参数名是msp名称 不是域名
		ccPolicy := cauthdsl.SignedByMspMember("FBIMSP")
		request := resmgmt.InstantiateCCRequest{Name: setup.ChainCodeID, Path: setup.ChaincodeGoPath, Version: setup.ChainCodeVersion, Args: [][]byte{[]byte("init"), []byte("init")}, Policy: ccPolicy}
		// opts := requestOptions{Targets: peers}
		resp, err := setup.admin.InstantiateCC(setup.ChannelID, request)
		if err != nil || resp.TransactionID == "" {
			return errors.WithMessage(err, "failed to instantiate the chaincode")
		}
		fmt.Println("Chaincode instantiate success")
	} else {
		fmt.Println("Chaincode has instantiated")
	}

	// Channel client is used to query and execute transactions
	clientContext := setup.sdk.ChannelContext(setup.ChannelID, fabsdk.WithUser(setup.UserName))
	// clientContext := setup.sdk.ChannelContext(setup.ChannelID, fabsdk.WithUser(setup.UserName), fabsdk.WithOrg("fbi.citizens.com"))
	setup.client, err = channel.New(clientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create new channel client")
	}
	fmt.Println("Channel client created")

	// Creation of the client which will enables access to our channel events
	setup.event, err = event.New(clientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create new event client")
	}
	fmt.Println("Event client created")

	fmt.Println("Chaincode Installation & Instantiation Successful")
	return nil
}

func (setup *FabricSetup) CloseSDK() {
	setup.sdk.Close()
}
