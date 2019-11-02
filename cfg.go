// cfg.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type lbElem struct {
	Host   string
	Weight uint32
}
type lbElemAry struct {
	Ary []lbElem
}

//{"Ary":[{"Host":"127.0.0.1:3301","Weight":8},{"Host":"127.0.0.1:3302","Weight":2}]}
var g_lbElemAry *lbElemAry

func GetCfgData(f string) *lbElemAry {
	if g_lbElemAry != nil {
		return g_lbElemAry
	}
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	fmt.Println("load cfg file:", string(dat))
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

	b, err := json.Marshal(lst)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	lbRes := lbElemAry{}
	err = json.Unmarshal(b, &lbRes)
	if err != nil {
		panic(err)
	}
	for _, elem := range lbRes.Ary {
		fmt.Println(elem.Host, elem.Weight)
	}
}
