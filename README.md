Ontology wasm
=====
## Introduction
Ontology wasm is a VM for ontology block chain, it can also be used for other stand-alone environment not only for block chains.

WebAssembly (abbreviated *Wasm*) is a binary instruction format for a stack-based virtual machine. Wasm is designed as a portable target for compilation of high-level languages like C/C++/Rust.



## Structure

Ontology-wasm disassemble and execute wasm binary codes based on [Wagon](https://github.com/go-interpreter/wagon) project with extra memory management. 

Currently, we support int, int64, float, double, string(byte array), int array and int64 array data types,since wasm only has 4 types (i32,i64,f32 and f64), that means only the 4 types data could be pushed into the stack, other complex data types must be stored in memory.

In wasm MVP version,every module can only has one linear memory.

![memory](./doc/images/memory.png)



## Useage

1. create a Engine to contain the VM, example like below:

```go
type ExecutionEngine struct {
	service *InteropService
	vm      *VM
	version  string //for test different contracts
	backupVM *vmstack
}

func NewExecutionEngine(iservice IInteropService, ver string) *ExecutionEngine {

	engine := &ExecutionEngine{
		service: NewInteropService(),
		version: ver,
	}
	if iservice != nil {
		engine.service.MergeMap(iservice.GetServiceMap())
	}

	engine.backupVM = newStack(VM_STACK_DEPTH)
	return engine
}
```
**service** contains the system apis which exists in the "import 'env' " section, that means you can create any api calls implemented by golang code.

We already put some apis in ```env_service.go```

**ver** represents the version of engine, you can use this field to decide how to deserialize your parameters.

Then load the wasm module(from a file or other stream)
```go
	code, err := ioutil.ReadFile("./test_data2/contract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}
```

Pass the Parameters (Json format for example)
```go
	par := make([]Param, 2)
	par[0] = Param{Ptype: "int", Pval: "20"}
	par[1] = Param{Ptype: "int", Pval: "30"}

	p := Args{Params: par}
	jbytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		t.Fatal(err.Error())
	}
	bf := bytes.NewBufferString("add")
	bf.WriteString("|")
	bf.Write(jbytes)

```
Execute the wasm vm 
```go
    	res, err := engine.Call(nil, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
```

If you know the result is not a basic type(int,int64,float or double),you should get the data from memory
```go
    	retbytes, err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " + string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes, result)
```

You can try the tests in ```exec/engine_test.go``` and smart contract tests in ```exec/contract_test.go```

## Ontology Smart contract
Please  refer to [smart-contract-tutorial](https://github.com/ontio/documentation/tree/master/smart-contract-tutorial)
