package main

import (
	"github.com/ipfs/go-cid"
	"os"
	"strings"
)
import "github.com/rs/xid"

//func CreateCid() (cid.Cid, error) {
//	pref := cid.Prefix{
//		Version:  1,
//		Codec:    cid.Raw,
//		MhType:   mh.SHA2_256,
//		MhLength: -1, // default length
//	}
//
//	// And then feed it some data
//	return pref.Sum([]byte("Hello World!"))
//}

func getEnvString() string {
	return strings.Join(os.Environ(), "")
}

func GetCidString(functionId string) string {
	c, err := cid.Decode(strings.Join([]string{getEnvString(), functionId}, ""))
	if err != nil {

	}
	return c.String()
}

func GenerateFunctionId() string {
	guid := xid.New()
	return guid.String()
}
