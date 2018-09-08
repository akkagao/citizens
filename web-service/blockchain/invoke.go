package blockchain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/akkagao/citizens/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// 注册用户
func (setup *FabricSetup) RegisterUser() (string, error) {

	// Prepare arguments
	var args []string

	args = append(args, "register")
	args = append(args, "535636789302345673")

	people := common.People{
		DataType:      "citizens",
		Id:            "535636789302345673",
		Sex:           "男",
		Name:          "张三",
		BirthLocation: common.Location{Province: "海南", City: "三亚市", Detail: "天涯海角"},
		LiveLocation:  common.Location{Province: "北京", Town: "朝阳区", Detail: "大悦城"},
		MotherId:      "535636789302345671",
		FatherId:      "535636789302345672",
		Childs:        []string{"535636789302345674", "535636789302345675"},
	}
	b, _ := json.Marshal(people)

	eventID := "eventInvoke"

	// Add data that will be visible in the proposal, like a description of the invoke request
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in hello invoke")

	reg, notifier, err := setup.event.RegisterChaincodeEvent(setup.ChainCodeID, eventID)
	if err != nil {
		return "", err
	}
	defer setup.event.Unregister(reg)

	// Create a request (proposal) and send it
	response, err := setup.client.Execute(channel.Request{
		ChaincodeID:  setup.ChainCodeID,
		Fcn:          args[0],
		Args:         [][]byte{[]byte(args[1]), b},
		TransientMap: transientDataMap})
	if err != nil {
		return "", fmt.Errorf("failed to move funds: %v", err)
	}

	// Wait for the result of the submission
	select {
	case ccEvent := <-notifier:
		fmt.Printf("Received CC event: %v\n", ccEvent)
	case <-time.After(time.Second * 20):
		return "", fmt.Errorf("did NOT receive CC event for eventId(%s)", eventID)
	}

	return string(response.TransactionID), nil
}
