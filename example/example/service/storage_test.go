package storage_test

import (
	"github.com/ontio/ontology-wasm/exec"
	"testing"
	"io/ioutil"
	"fmt"
	"encoding/binary"
	"bytes"
	"github.com/ontio/ontology_wasm_example/example/service"
	"github.com/ontio/ontology/common/serialization"
)

var service = exec.NewInteropService()

//合约中调用注册的方法，并操作内存获取参数并存储结果
func TestHelloRegisterContract(t *testing.T) {
	storage.Register(service)

	code, err := ioutil.ReadFile("../../data/storageContract1.wasm")
	if err != nil {
		t.Error("error in read file:", err.Error())
		return
	}

	bf := bytes.NewBufferString("get")
	bf.WriteString("|")

	serialization.WriteString(bf, "envin")

	//service provider strconcat
	engine := exec.NewExecutionEngine(service, "wasm_example")
	res, err := engine.Call(nil, code, bf.Bytes())
	if err != nil {
		t.Error("call error:", err.Error())
		return
	}
	fmt.Printf("res:%v \n", res)
	fmt.Println(string(engine.GetMemory().Memory))

	retbytes, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " + string(retbytes))
}


func TestCallContract (t *testing.T) {

}

