package utils

import (
	"golang.org/x/crypto/ripemd160"
	"encoding/hex"
)

func GenContractAddress (code []byte) string{
	rip := ripemd160.New()
	rip.Write(code)
	address := rip.Sum(nil)
	return hex.EncodeToString(address)
}