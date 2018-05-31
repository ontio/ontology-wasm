package storage

import (
	"github.com/ontio/ontology-wasm/exec"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
	"github.com/ontio/ontology-wasm/memory"
	"fmt"
	"github.com/ontio/ontology-wasm/example/utils"
)

func Register(service *exec.InteropService) {
	service.Register("strconcat", strconcat)
	service.Register("putStorage", putStorage)
	service.Register("getStorage", getStorage)
	service.Register("log", log)
}

func strconcat(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 2 {
		return false, errors.New("[putStorage] parameter count error")
	}
	firstParam, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	secondParam, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, err
	}
	res := make([]byte, len(firstParam)+len(secondParam))
	copy(res[0:len(firstParam)], firstParam)
	copy(res[len(firstParam):], secondParam)
	if point, err := vm.SetMemory(string(res)); err != nil {
		return false, err
	} else {
		vm.RestoreCtx()
		vm.PushResult(uint64(point))
	}

	return true, nil
}

func putStorage(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 2 {
		return false, errors.New("[putStorage] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	if len(key) > 1024 {
		return false, errors.New("[putStorage] Get Storage key to long")
	}

	value, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, err
	}
	save(key,value)

	vm.RestoreCtx()

	return true, nil
}

func save (key interface{},value interface{}) {
	coll := utils.GetMongoDB().C("contract_data")
	if count, err := coll.Find(bson.M{"key": key}).Count(); err != nil {
		fmt.Printf("save data error:%s\n",err.Error())
	} else {
		if count >= 1 {
			coll.Update(bson.M{
				"key": key,
			}, bson.M{
				"$set": bson.M{"value": value},
			})
		} else {
			coll.Insert(bson.M{
				"key":   key,
				"value": value,
			})
		}
	}
}

func getStorage(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.New("[getStorage] parameter count error ")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	coll := utils.GetMongoDB().C("contract_data")
	m := make(map[string]interface{})
	coll.Find(bson.M{
		"key": key,
	}).Select(bson.M{"_id": false, "value": true}).One(m)
	if len(m) == 0 {
		vm.RestoreCtx()
		if envCall.GetReturns() {
			vm.PushResult(uint64(memory.VM_NIL_POINTER))
		}
		return true, nil
	}
	index, err := vm.SetPointerMemory(m["value"])
	if err != nil {
		return false, nil
	}

	vm.RestoreCtx()
	if envCall.GetReturns() {
		vm.PushResult(uint64(index))
	}
	return true, nil
}

func log(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	for _,param := range params {
		point, err := vm.GetPointerMemory(param)
		if err != nil {
			return false, err
		}
		fmt.Print(string(point))
	}
	fmt.Println()

	vm.RestoreCtx()
	return true, nil
}
