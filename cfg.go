package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type lbElem struct {
	Host   string
	Weight uint32
}
type lbElemAry struct {
	Ary  []lbElem
	Port string
    LogDir string
}

//{"Port":"3344","Ary":[{"Host":"127.0.0.1:12335","Weight":8},{"Host":"127.0.0.1:12336","Weight":2}]}
var g_lbElemAry *lbElemAry

func GetCfgData(f string) *lbElemAry {
	if g_lbElemAry != nil {
		return g_lbElemAry
	}
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	log.Printf("load cfg file: %s", string(dat))
	g_lbElemAry = &lbElemAry{}
	err = json.Unmarshal(dat, g_lbElemAry)
	if err != nil {
		panic(err)
	}
	return g_lbElemAry
}

func test_cfg() {

	lst := lbElemAry{}

	e := lbElem{}
	e.Host = "127.0.0.1:3301"
	e.Weight = uint32(8)
	lst.Ary = append(lst.Ary, e)

	e = lbElem{}
	e.Host = "127.0.0.1:3302"
	e.Weight = uint32(2)
	lst.Ary = append(lst.Ary, e)
	lst.Port = "3344"

	b, err := json.Marshal(lst)
	if err != nil {
		panic(err)
	}
	log.Printf(string(b))

	lbRes := lbElemAry{}
	err = json.Unmarshal(b, &lbRes)
	if err != nil {
		panic(err)
	}
	for _, elem := range lbRes.Ary {
		log.Printf(elem.Host, elem.Weight)
	}
}
