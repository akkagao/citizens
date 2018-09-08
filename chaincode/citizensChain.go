package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

/**
个人基础信息
*/
type People struct {
	DataType      string   `json:"dataType"`      // 区分数据类型
	Id            string   `json:"id"`            // 身份证号码
	Sex           string   `json:"sex"`           // 性别
	Name          string   `json:"name"`          // 姓名
	BirthLocation Location `json:"birthLocation"` // 出生地
	LiveLocation  Location `json:"liveLocation"`  // 现居住地
	MotherId      string   `json:"motherID"`      // 母亲身份证号码
	FatherId      string   `json:"fatherID"`      // 父亲身份证号码
	Childs        []string `json:"chailds"`       // 子女身份证
}

/**
位置
*/
type Location struct {
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 城市
	Town     string `json:"town"`     // 镇
	Detail   string `json:"detail"`   // 详细住址
}

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

	if args[0] != "init" {
		return shim.Error("args error")
	}
	log.Println("Init CitizensChain success")
	return shim.Success(nil)
}

/**
执行查询、插入等方法
*/
func (c *CitizensChain) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	log.Println("========Invoke========")
	function, args := stub.GetFunctionAndParameters()
	log.Println("========GetFunctionAndParameters========", function, len(args), args)
	log.Println(function == "register")
	log.Println(function == "query")

	if function == "register" {
		log.Println("========register========")
		return c.register(stub, args)
	} else if function == "query" {
		log.Println("========query========")
		return c.query(stub, args)
	} else {
		return shim.Error("function not define hahah")
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

	people := People{}
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
