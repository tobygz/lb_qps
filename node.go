package main

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Node struct {
	Host   string
	StartW uint32
	EndW   uint32
	Weight uint32

	conn  net.Conn
	bconn bool
}

func (self *Node) init() bool {
	var err error
	self.conn, err = net.Dial("tcp", self.Host)
	if err != nil {
		log.Printf("net dial host fail: %v", self.Host)
		return false
	}
	log.Printf("connect to %s succ", self.Host)
	self.bconn = true
	return true

}

func (self *Node) Dowork(pbp *PBDataPack) *PBDataPack {
	log.Printf("send to %s", self.Host)
	if pbp.Send(self.conn) == false {
		self.bconn = false
		return nil
	}
	ret := &PBDataPack{}
	if ret.Unpack(self.conn) == false {
		self.bconn = false
		return nil
	}
	return ret
}

type NodeList struct {
	_lst         []*Node
	_totalWeight uint32
	exitCh       chan bool
	sync.Mutex
}

var g_NodeList *NodeList

func GetNodeList() *NodeList {
	if g_NodeList != nil {
		return g_NodeList
	}
	g_NodeList = &NodeList{
		_lst:         make([]*Node, 0),
		_totalWeight: uint32(0),
		exitCh:       make(chan bool, 1),
	}

	return g_NodeList
}

func (self *NodeList) _doRebalance() {
	//get _totalWeight
	self._totalWeight = uint32(0)
	for _, node := range self._lst {
		if node.bconn == false {
			continue
		}
		self._totalWeight += node.Weight
	}

	//calc startw & endw
	startW := uint32(0)
	for _, node := range self._lst {
		if node.bconn == false {
			continue
		}
		node.StartW = startW
		node.EndW = startW + node.Weight
		startW = node.EndW
	}
	log.Printf("after dorebal total: %d, alivecount: %d", self._totalWeight, self._aliveCount())
}

func (self *NodeList) init(cfgf string) {
	rand.Seed(time.Now().UnixNano())

	//born all node from cfg
	cfg := GetCfgData(cfgf)
	self._totalWeight = uint32(0)
	for _, elem := range cfg.Ary {
		self._totalWeight += elem.Weight
	}

	startW := uint32(0)
	for _, elem := range cfg.Ary {
		node := &Node{
			bconn: false,
		}
		node.Host = elem.Host
		node.StartW = startW
		node.EndW = startW + elem.Weight
		node.Weight = elem.Weight
		self._lst = append(self._lst, node)
		startW = node.EndW
	}

	//connect to all host
	bfind := false
	for _, node := range self._lst {
		if !node.init() {
			bfind = true
		}
	}
	if bfind {
		self._doRebalance()
	}

	//init timer, do reconnect
	self._doReconnect()
}

func (self *NodeList) _doReconnect() {
	tk := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-tk.C:
				self.ChkAlive()
			case <-self.exitCh:
				return
			}
		}
	}()
}

func (self *NodeList) _aliveCount() int {
	ct := 0
	for _, nd := range self._lst {
		if nd.bconn {
			ct++
		}
	}
	return ct
}

func (self *NodeList) ChkAlive() {
	self.Lock()
	defer self.Unlock()
	bfind := false
	for _, nd := range self._lst {
		if nd.bconn == true {
			continue
		}
		if nd.init() {
			bfind = true
		}
	}
	if bfind {
		self._doRebalance()
	}
}

func (self *NodeList) Dispatch(pbp *PBDataPack) *PBDataPack {
	self.Lock()
	x := rand.Intn(int(self._totalWeight))
	log.Printf("Dispatch total: %d ,nowrand: %d", self._totalWeight, x)

	for _, nd := range self._lst {
		if nd.bconn == false {
			continue
		}
		if uint32(x) >= nd.StartW && uint32(x) < nd.EndW {
			ret := nd.Dowork(pbp)
			if ret != nil {
				self.Unlock()
				return ret
			}
		}
	}
	self.Unlock()
	//not get at all
	if self._aliveCount() > 0 {
		self._doRebalance()
		return self.Dispatch(pbp)
	}
	log.Printf("Dispatch fail, no alived server")
	return nil
}
