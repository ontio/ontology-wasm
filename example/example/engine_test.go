package example_test

import (
	"github.com/ontio/ontology-wasm/exec"
	"testing"
	"io/ioutil"
	"fmt"
	"encoding/binary"
	"encoding/json"
	"bytes"
	"github.com/ontio/ontology/common/serialization"
)

var service = exec.NewInteropService()

func TestHello(t *testing.T) {
	code, err := ioutil.ReadFile("../data/hello.wasm")
	if err != nil {
		t.Error("error in read file:", err.Error())
		return
	}

	method := "hello"
	name := "envin"
	input := make([]byte, 1+len(method)+1+1+len(name))

	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(1)            //parameters  count
	input[len(method)+2] = byte(len(name))    //first parameters length
	copy(input[len(method)+3:], []byte(name)) //parameter

	fmt.Printf("input:%v \n", input)

	input1 := make([]interface{}, 2)
	input1[0] = "hello"
	input1[1] = "envin"

	//contact provider strconcat
	engine := exec.NewExecutionEngine(service, "test")
	//result,err := engine.CallInf(nil,code,input1,nil)
	result, err := engine.Call(nil, code, input)
	if err != nil {
		t.Error("call error:", err.Error())
		return
	}

	fmt.Printf("result:%v \n", result)
	offset := uint64(binary.LittleEndian.Uint32(result))
	reBs, err := engine.GetMemory().GetPointerMemory(offset)
	if err != nil {
		t.Errorf("get memory error:%s", err.Error())
	} else {
		fmt.Printf("result:%s \n", string(reBs))
	}
}

//普通方法，并注册strconcat方法
func TestHelloRegister(t *testing.T) {
	//register srconcat
	service.Register("strconcat", func(engine *exec.ExecutionEngine) (bool, error) {
		mem := engine.GetMemory().Memory
		params := engine.GetVM().GetEnvCall().GetParams()
		str := string(mem[params[0]:params[0]+params[1]])
		engine.GetVM().RestoreCtx()
		engine.GetVM().PushResult(uint64(len(str)))
		return true, nil
	})

	code, err := ioutil.ReadFile("../data/hello1.wasm")
	if err != nil {
		t.Error("error in read file:", err.Error())
		return
	}

	input1 := make([]interface{}, 2)
	input1[0] = "hello"
	input1[1] = "envin"

	//service provider strconcat
	engine := exec.NewExecutionEngine(service, "wasm_example")
	result, err := engine.CallInf(nil, code, input1, nil)
	if err != nil {
		t.Error("call error:", err.Error())
		return
	}
	fmt.Printf("result:%v \n", result)
	if err != nil {
		t.Errorf("get memory error:%s", err.Error())
	} else {
		fmt.Println("result:", string(engine.GetMemory().Memory[:]))
	}
}

//测试hello的合约调用
func TestHelloContract(t *testing.T) {
	code, err := ioutil.ReadFile("../data/helloContract.wasm")
	if err != nil {
		t.Error("error in read file:", err.Error())
		return
	}

	par := make([]exec.Param, 1)
	par[0] = exec.Param{Ptype: "string", Pval: "envin"}

	p := exec.Args{Params: par}
	input, _ := json.Marshal(p)
	fmt.Printf("param:%s \n", string(input))

	bf := bytes.NewBufferString("hello")
	bf.WriteString("|")
	//bf.WriteString("envin")

	//serialization.WriteString(bf,"|")
	//在合约中调用了service的ReadStringParam函数，这个函数只能读取被serialization序列化过的参数
	serialization.WriteString(bf, "envin")

	fmt.Printf("input:%s \n", bf.String())
	fmt.Printf("input:%v \n", bf.Bytes())

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

//合约中调用注册的方法，并操作内存获取参数并存储结果
func TestHelloRegisterContract(t *testing.T) {
	//register srconcat
	service.Register("strconcat", func(engine *exec.ExecutionEngine) (bool, error) {
		params := engine.GetVM().GetEnvCall().GetParams()
		firstParam, err := engine.GetMemory().GetPointerMemory(params[0])
		if err != nil {
			return false, err
		}
		secondParam, err := engine.GetMemory().GetPointerMemory(params[1])
		if err != nil {
			return false, err
		}
		res := make([]byte, len(firstParam)+len(secondParam))
		copy(res[0:len(firstParam)], firstParam)
		copy(res[len(firstParam):], secondParam)
		if point, err := engine.GetVM().SetMemory(string(res)); err != nil {
			return false, err
		} else {
			engine.GetVM().RestoreCtx()
			engine.GetVM().PushResult(uint64(point))
		}

		return true, nil
	})

	code, err := ioutil.ReadFile("../data/helloContract2.wasm")
	if err != nil {
		t.Error("error in read file:", err.Error())
		return
	}

	par := make([]exec.Param, 1)
	par[0] = exec.Param{Ptype: "string", Pval: "envin"}

	p := exec.Args{Params: par}
	input, _ := json.Marshal(p)
	fmt.Printf("param:%s \n", string(input))

	bf := bytes.NewBufferString("hello")
	bf.WriteString("|")

	serialization.WriteString(bf, "envin")

	fmt.Printf("input:%s \n", bf.String())
	fmt.Printf("input:%v \n", bf.Bytes())
	//bs :=

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

