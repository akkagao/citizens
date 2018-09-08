package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/akkagao/citizens/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// 查询用户
func (setup *FabricSetup) QueryUser() (string, error) {

	var args []string
	// 参数1 作为调用Invoke方法的function 参数
	args = append(args, "query")
	// 调用Invoke 的参数
	args = append(args, "535636789302345673")

	response, err := setup.client.Query(channel.Request{
		ChaincodeID: setup.ChainCodeID,
		Fcn:         args[0],
		Args:        [][]byte{[]byte(args[1])},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query: %v", err)
	}

	people := common.People{}
	err = json.Unmarshal([]byte(response.Payload), &people)
	fmt.Println(people)

	return string(response.Payload), nil
}
