package utils_test

import (
	"testing"
	"io/ioutil"
	"fmt"
	"github.com/ontio/ontology_wasm_example/utils"
)

func TestRipemd160(t *testing.T) {
	code,err := ioutil.ReadFile("../data/hello.wasm")
	if err != nil {
		fmt.Println("read file error:",err.Error())
	}else {
		address := utils.GenContractAddress(code)
		fmt.Println(address)
	}
}
