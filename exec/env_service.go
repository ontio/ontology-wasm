/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package exec

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/ontio/ontology-wasm/memory"
	"github.com/ontio/ontology/common/serialization"
)

type Args struct {
	Params []Param `json:"Params"`
}

type Param struct {
	Ptype string `json:"type"`
	Pval  string `json:"value"`
}

type Result struct {
	Ptype string `json:"type"`
	Pval  string `json:"value"`
}

type InteropServiceInterface interface {
	Register(method string, handler func(*ExecutionEngine) (bool, error)) bool
	GetServiceMap() map[string]func(*ExecutionEngine) (bool, error)
}

type InteropService struct {
	serviceMap map[string]func(*ExecutionEngine) (bool, error)
}

func NewInteropService() *InteropService {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	//init some system functions
	service.Register("calloc", calloc)
	service.Register("strcmp", stringcmp)
	service.Register("malloc", malloc)
	service.Register("arrayLen", arrayLen)
	service.Register("memcpy", memcpy)
	service.Register("memset", memset)
	service.Register("getTempRet0", getTempRet0)


	service.Register("ReadInt32Param", readInt32Param)
	service.Register("ReadInt64Param", readInt64Param)
	service.Register("ReadStringParam", readStringParam)
	service.Register("JsonUnmashalInput", jsonUnmashal)
	service.Register("JsonMashalResult", jsonMashal)
	service.Register("JsonMashalParams", jsonMashalParams)
	service.Register("RawMashalParams", rawMashalParams)


	//===================add block apis below==================
	return &service
}

func (i *InteropService) Register(name string, handler func(*ExecutionEngine) (bool, error)) bool {
	if _, ok := i.serviceMap[name]; ok {
		return false
	}
	i.serviceMap[name] = handler
	return true
}

func (i *InteropService) Invoke(methodName string, engine *ExecutionEngine) (bool, error) {

	if v, ok := i.serviceMap[methodName]; ok {
		return v(engine)
	}
	return false, errors.New("Not supported method:" + methodName)
}

func (i *InteropService) MergeMap(mMap map[string]func(*ExecutionEngine) (bool, error)) bool {

	for k, v := range mMap {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
	return true
}

func (i *InteropService) Exists(name string) bool {
	_, ok := i.serviceMap[name]
	return ok
}
func (i *InteropService) GetServiceMap() map[string]func(*ExecutionEngine) (bool, error) {
	return i.serviceMap
}

//******************* basic functions ***************************
//TODO decide to replace the P_UNKNOW type

//for the c language "calloc" function
func calloc(engine *ExecutionEngine) (bool, error) {

	envCall := engine.vm.envCall
	params := envCall.envParams

	if len(params) != 2 {
		return false, errors.New("parameter count error while call calloc")
	}
	count := int(params[0])
	length := int(params[1])
	//we don't know whats the alloc type here
	index, err := engine.vm.memory.MallocPointer(count*length, memory.PUnkown)
	if err != nil {
		return false, err
	}

	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(index))
	}
	return true, nil
}

//for the c language "malloc" function
func malloc(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call calloc")
	}
	size := int(params[0])
	//we don't know whats the alloc type here
	index, err := engine.vm.memory.MallocPointer(size, memory.PUnkown)
	if err != nil {
		return false, err
	}
	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(index))
	}
	return true, nil

}

//use arrayLen to replace 'sizeof'
func arrayLen(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call arrayLen")
	}

	pointer := params[0]

	tl, ok := engine.vm.memory.MemPoints[pointer]

	var result uint64
	if ok {
		switch tl.Ptype {
		case memory.PInt8, memory.PString:
			result = uint64(tl.Length / 1)
		case memory.PInt16:
			result = uint64(tl.Length / 2)
		case memory.PInt32, memory.PFloat32:
			result = uint64(tl.Length / 4)
		case memory.PInt64, memory.PFloat64:
			result = uint64(tl.Length / 8)
		case memory.PUnkown:
			//todo assume it's byte
			result = uint64(tl.Length / 1)
		default:
			result = uint64(0)
		}

	} else {
		result = uint64(0)
	}
	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(result))
	}
	return true, nil

}

func memcpy(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 3 {
		return false, errors.New("parameter count error while call memcpy")
	}
	dest := int(params[0])
	src := int(params[1])
	length := int(params[2])

	if dest < src && dest+length > src {
		return false, errors.New("memcpy overlapped")
	}

	copy(engine.vm.memory.Memory[dest:dest+length], engine.vm.memory.Memory[src:src+length])

	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(1))
	}

	return true, nil //this return will be dropped in wasm
}

func memset(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 3 {
		return false, errors.New("parameter count error while call memcpy")
	}
	dest := int(params[0])
	char := int(params[1])
	cnt := int(params[2])

	tmp := make([]byte, cnt)
	for i := 0; i < cnt; i++ {
		tmp[i] = byte(char)
	}
	copy(engine.vm.Memory()[dest:dest+cnt], tmp)

	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(1))
	}

	return true, nil //this return will be dropped in wasm
}

func getTempRet0(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall

	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(0))
	}
	return true, nil //this return will be dropped in wasm
}

//read int value from args bytes
func readInt32Param(engine *ExecutionEngine) (bool, error) {

	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call readInt32Param")
	}

	addr := params[0]
	paramBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, err
	}

	pidx := engine.vm.memory.ParamIndex

	if pidx+4 > len(paramBytes) {
		return false, errors.New("read params error")
	}

	retInt := binary.LittleEndian.Uint32(paramBytes[pidx : pidx+4])
	engine.vm.memory.ParamIndex += 4

	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(retInt))
	}
	return true, nil
}

//read int64 value from args bytes
func readInt64Param(engine *ExecutionEngine) (bool, error) {

	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call readInt64Param")
	}
	addr := params[0]
	paramBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, err
	}
	pidx := engine.vm.memory.ParamIndex
	if pidx+8 > len(paramBytes) {
		return false, errors.New("read params error")
	}

	retInt := binary.LittleEndian.Uint64(paramBytes[pidx : pidx+8])
	engine.vm.memory.ParamIndex += 8

	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(retInt)
	}
	return true, nil
}

//read string value from args bytes
func readStringParam(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call readStringParam")
	}

	addr := params[0]
	paramBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, err
	}
	var length int
	pidx := engine.vm.memory.ParamIndex
	switch paramBytes[pidx] {
	case 0xfd: //uint16
		if pidx+3 > len(paramBytes) {
			return false, errors.New("read string failed")
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pidx+1 : pidx+3]))
		pidx += 3
	case 0xfe: //uint32
		if pidx+5 > len(paramBytes) {
			return false, errors.New("read string failed")
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pidx+1 : pidx+5]))
		pidx += 5
	case 0xff:
		if pidx+9 > len(paramBytes) {
			return false, errors.New("read string failed")
		}
		length = int(binary.LittleEndian.Uint16(paramBytes[pidx+1 : pidx+9]))
		pidx += 9
	default:
		length = int(paramBytes[pidx])
	}

	if pidx+length > len(paramBytes) {
		return false, errors.New("read string failed")
	}
	pidx += length + 1

	stringbytes := paramBytes[engine.vm.memory.ParamIndex+1 : pidx]

	retidx, err := engine.vm.SetPointerMemory(stringbytes)
	if err != nil {
		return false, errors.New("set memory failed")
	}

	engine.vm.memory.ParamIndex = pidx
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(retidx))
	}
	return true, nil
}

func jsonUnmashal(engine *ExecutionEngine) (bool, error) {

	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 3 {
		return false, errors.New("parameter count error while call jsonUnmashal")
	}

	addr := params[0]
	size := int(params[1])

	jsonaddr := params[2]
	jsonbytes, err := engine.vm.GetPointerMemory(jsonaddr)
	if err != nil {
		return false, err
	}
	arg := &Args{}
	err = json.Unmarshal(jsonbytes, arg)

	if err != nil {
		return false, err
	}

	buff := bytes.NewBuffer(nil)

	for _, arg := range arg.Params {

		switch strings.ToLower(arg.Ptype) {
		case "int":
			tmp := make([]byte, 4)
			val, err := strconv.Atoi(arg.Pval)
			if err != nil {
				return false, err
			}
			binary.LittleEndian.PutUint32(tmp, uint32(val))
			buff.Write(tmp)

		case "int64":
			tmp := make([]byte, 8)
			val, err := strconv.ParseInt(arg.Pval, 10, 64)
			if err != nil {
				return false, err
			}
			binary.LittleEndian.PutUint64(tmp, uint64(val))
			buff.Write(tmp)

		case "int_array":
			arr := strings.Split(arg.Pval, ",")
			tmparr := make([]int, len(arr))
			for i, str := range arr {
				tmparr[i], err = strconv.Atoi(str)
				if err != nil {
					return false, err
				}
			}
			idx, err := engine.vm.SetPointerMemory(tmparr)
			if err != nil {
				return false, err
			}
			tmp := make([]byte, 4)
			binary.LittleEndian.PutUint32(tmp, uint32(idx))
			buff.Write(tmp)

		case "int64_array":
			arr := strings.Split(arg.Pval, ",")
			tmparr := make([]int64, len(arr))
			for i, str := range arr {
				tmparr[i], err = strconv.ParseInt(str, 10, 64)
				if err != nil {
					return false, err
				}
			}

			idx, err := engine.vm.SetPointerMemory(tmparr)
			if err != nil {
				return false, err
			}
			tmp := make([]byte, 8)
			binary.LittleEndian.PutUint64(tmp, uint64(idx))
			buff.Write(tmp)

		case "string":
			idx, err := engine.vm.SetPointerMemory(arg.Pval)
			if err != nil {
				return false, err
			}
			tmp := make([]byte, 4)
			binary.LittleEndian.PutUint32(tmp, uint32(idx))
			buff.Write(tmp)

		default:
			return false, errors.New("unsupported type :" + arg.Ptype)
		}

	}

	bytes := buff.Bytes()
	if len(bytes) != size {
		//return false ,errors.New("")
		//todo this case is not an error, sizeof doesn't means actual memory length,so the size parameter should be removed.
		//fmt.Printf("length is not same! size :%d, length:%d\n", size, len(bytes))
	}

	//todo move to SetMemory method
	engine.vm.Malloc(len(bytes))
	copy(engine.vm.memory.Memory[int(addr):int(addr)+len(bytes)], bytes)

	engine.vm.RestoreCtx()
	return true, nil
}

func jsonMashal(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams

	if len(params) != 2 {
		return false, errors.New("parameter count error while call jsonUnmashal")
	}

	val := params[0]
	ptype := params[1] //type
	tpstr, err := engine.vm.GetPointerMemory(ptype)
	if err != nil {
		return false, err
	}

	ret := &Result{}

	pstype := strings.ToLower(TrimBuffToString(tpstr))
	ret.Ptype = pstype
	switch pstype {
	case "int":
		res := int(val)
		ret.Pval = strconv.Itoa(res)

	case "int64":
		res := int64(val)
		ret.Pval = strconv.FormatInt(res, 10)

	case "string":
		tmp, err := engine.vm.GetPointerMemory(val)
		if err != nil {
			return false, err
		}
		ret.Pval = TrimBuffToString(tmp)

	case "int_array":
		tmp, err := engine.vm.GetPointerMemory(val)
		if err != nil {
			return false, err
		}
		length := len(tmp) / 4
		retArray := make([]string, length)
		for i := 0; i < length; i++ {
			retArray[i] = strconv.Itoa(int(binary.LittleEndian.Uint32(tmp[i : i+4])))
		}
		ret.Pval = strings.Join(retArray, ",")

	case "int64_array":
		tmp, err := engine.vm.GetPointerMemory(val)
		if err != nil {
			return false, err
		}
		length := len(tmp) / 8
		retArray := make([]string, length)
		for i := 0; i < length; i++ {
			retArray[i] = strconv.FormatInt(int64(binary.LittleEndian.Uint64(tmp[i:i+8])), 10)
		}
		ret.Pval = strings.Join(retArray, ",")
	}

	jsonstr, err := json.Marshal(ret)
	if err != nil {
		return false, err
	}

	offset, err := engine.vm.SetPointerMemory(string(jsonstr))
	if err != nil {
		return false, err
	}

	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(offset))
	}

	return true, nil
}

func stringcmp(engine *ExecutionEngine) (bool, error) {

	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 2 {
		return false, errors.New("parameter count error while call strcmp")
	}

	var ret int

	addr1 := params[0]
	addr2 := params[1]
	if addr1 == addr2 {
		ret = 0
	} else {
		bytes1, err := engine.vm.GetPointerMemory(addr1)
		if err != nil {
			return false, err
		}

		bytes2, err := engine.vm.GetPointerMemory(addr2)
		if err != nil {
			return false, err
		}

		if TrimBuffToString(bytes1) == TrimBuffToString(bytes2) {
			ret = 0
		} else {
			ret = 1
		}
	}
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(ret))
	}
	return true, nil
}


func jsonMashalParams(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("[jsonMashalParams]parameter count error")
	}

	addr := params[0]

	bytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
	}

	var pars []Param

	bytesLen := len(bytes)
	for i := 0; i < bytesLen; i++ {
		typeIdx := binary.LittleEndian.Uint32(bytes[i : i+4])
		typeBytes, err := engine.vm.GetPointerMemory(uint64(typeIdx))
		if err != nil {
			return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
		}

		sType := strings.ToLower(TrimBuffToString(typeBytes))

		param := Param{Ptype: sType}

		switch sType {
		case "int":
			intBytes := int(binary.LittleEndian.Uint32(bytes[i+4 : i+8]))
			param.Pval = strconv.Itoa(intBytes)
			i += 7
		case "int64":
			intBytes := int64(binary.LittleEndian.Uint64(bytes[i+4 : i+12]))
			param.Pval = strconv.FormatInt(intBytes, 10)
			i += 11
		case "string":
			intBytes := uint64(binary.LittleEndian.Uint32(bytes[i+4 : i+8]))
			str, err := engine.vm.GetPointerMemory(intBytes)
			if err != nil {
				return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
			}
			param.Pval = TrimBuffToString(str)
			i += 7
		default:
			return false, errors.New("[jsonMashalParams]  not support type :" + string(typeBytes))
		}

		pars = append(pars, param)
	}

	arg := &Args{Params: pars}
	argJson, err := json.Marshal(arg)
	if err != nil {
		return false, errors.New("[jsonMashalParams] json.Marshal err:" + err.Error())
	}

	argIdx, err := engine.vm.SetPointerMemory(argJson)
	if err != nil {
		return false, errors.New("[jsonMashalParams] SetPointerMemory err:" + err.Error())
	}
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(argIdx))
	}

	return true, nil

}

func rawMashalParams(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("[jsonMashalParams]parameter count error")
	}

	addr := params[0]

	pBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
	}
	bf := bytes.NewBuffer(nil)

	bytesLen := len(pBytes)
	for i := 0; i < bytesLen; i++ {
		typeIdx := binary.LittleEndian.Uint32(pBytes[i : i+4])
		typeBytes, err := engine.vm.GetPointerMemory(uint64(typeIdx))
		if err != nil {
			return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
		}

		sType := strings.ToLower(TrimBuffToString(typeBytes))

		switch sType {
		case "int":
			intBytes := binary.LittleEndian.Uint32(pBytes[i+4 : i+8])
			tmpBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(tmpBytes, intBytes)
			bf.Write(tmpBytes)
			i += 7
		case "int64":
			intBytes := binary.LittleEndian.Uint64(pBytes[i+4 : i+12])
			tmpBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(tmpBytes, intBytes)
			bf.Write(tmpBytes)
			i += 11
		case "string":
			intBytes := uint64(binary.LittleEndian.Uint32(pBytes[i+4 : i+8]))
			str, err := engine.vm.GetPointerMemory(intBytes)
			if err != nil {
				return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
			}

			tmp := bytes.NewBuffer(nil)
			serialization.WriteString(tmp, TrimBuffToString(str))
			bf.Write(tmp.Bytes())
			i += 7
		default:
			return false, errors.New("[jsonMashalParams]  not support type :" + string(typeBytes))
		}

	}

	argIdx, err := engine.vm.SetPointerMemory(bf.Bytes())
	if err != nil {
		return false, errors.New("[jsonMashalParams] SetPointerMemory err:" + err.Error())
	}
	engine.vm.RestoreCtx()
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(argIdx))
	}
	return true, nil

}
//trim the '\00' byte
func TrimBuffToString(bytes []byte) string {

	for i, b := range bytes {
		if b == 0 {
			return string(bytes[:i])
		}
	}
	return string(bytes)

}
