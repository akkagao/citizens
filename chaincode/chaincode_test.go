package main

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var chaincodeName = "CitizensChain"

// chaincode_example05 looks like it wanted to return a JSON response to Query()
// it doesn't actually do this though, it just returns the sum value
func jsonResponse(name string, value string) string {
	return fmt.Sprintf("jsonResponse = \"{\"Name\":\"%v\",\"Value\":\"%v\"}", name, value)
}

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, name string, expect string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != expect {
		fmt.Println("State value", name, "was not", expect, "as expected")
		t.FailNow()
	}
}

func checkQuery(t *testing.T, stub *shim.MockStub, args [][]byte, expect string) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Query", args, "failed", string(res.Message))
		t.FailNow()
	}
	if res.Payload == nil {
		fmt.Println("Query", args, "failed to get result")
		t.FailNow()
	}

	if string(res.Payload) != expect {
		fmt.Println("Query result ", string(res.Payload), "was not", expect, "as expected")
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
}
func TestInit(t *testing.T) {
	scc := new(CitizensChain)
	stub := shim.NewMockStub("CitizensChain", scc)

	checkInit(t, stub, [][]byte{[]byte("init"), []byte("citizensChain")})
}

func TestQuery(t *testing.T) {
	scc := new(CitizensChain)
	stub := shim.NewMockStub("CitizensChain", scc)

	ccEx2 := new(CitizensChain)
	stubEx2 := shim.NewMockStub(chaincodeName, ccEx2)
	checkInit(t, stubEx2, [][]byte{[]byte("init"), []byte("a"), []byte("111"), []byte("b"), []byte("222")})
	stub.MockPeerChaincode(chaincodeName, stubEx2)

	checkInit(t, stub, [][]byte{[]byte("init"), []byte("sumStoreName"), []byte("0")})

	// a + b = 111 + 222 = 333
	checkQuery(t, stub, [][]byte{[]byte("query"), []byte(chaincodeName), []byte("sumStoreName"), []byte("")}, "333") // example05 doesn't return JSON?
}

func TestRegister(t *testing.T) {
	scc := new(CitizensChain)
	stub := shim.NewMockStub("CitizensChain", scc)
	checkInit(t, stub, [][]byte{[]byte("init"), []byte("citizensChain")})
	people := People{
		DataType:      "citizens",
		Id:            "535636789302345673",
		Sex:           "男",
		Name:          "张三",
		BirthLocation: Location{Province: "海南", City: "三亚市", Detail: "天涯海角"},
		LiveLocation:  Location{Province: "北京", Town: "朝阳区", Detail: "大悦城"},
		MotherId:      "535636789302345671",
		FatherId:      "535636789302345672",
		Childs:        []string{"535636789302345674", "535636789302345675"},
	}
	b, _ := json.Marshal(people)
	checkInvoke(t, stub, [][]byte{[]byte("register"), []byte("535636789302345673"), b})
	checkQuery(t, stub, [][]byte{[]byte("query"), []byte("535636789302345673")}, string(b))
}
