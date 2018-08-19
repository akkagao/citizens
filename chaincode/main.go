package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/akkagao/citizens/common"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

/**
公民链
存储公民基础信息
*/
type CitizensChain struct {
}

/**
初始化方法
*/
func (c *CitizensChain) Init(stub shim.ChaincodeStubInterface) peer.Response {
	log.Println("Init CitizensChain start")
	function, args := stub.GetFunctionAndParameters()
	if function != "init" {
		return shim.Error("function is not define")
	}
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	if args[0] != "citizensChain" {
		return shim.Error("args error")
	}
	log.Println("Init CitizensChain success")
	return shim.Success(nil)
}

/**
执行查询、插入等方法
*/
func (c *CitizensChain) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "register" {
		return c.register(stub, args)
	} else if function == "query" {
		return c.query(stub, args)
	} else {
		return shim.Error("function not define")
	}
	return shim.Success(nil)
}

/**
查询公民信息
*/
func (c *CitizensChain) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) < 1 {
		return shim.Error("args error")
	}
	key := args[0]
	result, err := stub.GetState(key)
	if err != nil {
		log.Println(fmt.Sprintf("query fail key:%s err:%s", key, err))
		return shim.Error("query fail")
	}
	return shim.Success(result)
}

/**
录入公民信息
*/
func (c *CitizensChain) register(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("args error")
	}
	// 身份证号码
	key := args[0]
	// 公民信息（用json保存）
	value := args[1]

	people := common.People{}
	err := json.Unmarshal([]byte(value), &people)
	if err != nil {
		return shim.Error("register fail, parameters cannot be parsed into json objects")
	}
	stub.PutState(key, []byte(value))
	log.Println(key, args)
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(CitizensChain))
	if err != nil {
		log.Println(err)
	}

}
