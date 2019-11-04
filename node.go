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

	//conn  net.Conn
	connChan chan net.Conn
	aliveCt  int
	sync.Mutex
}

func (self *Node) incAlive() {
	self.Lock()
	defer self.Unlock()
	self.aliveCt++
}
func (self *Node) GetAlive() int {
	self.Lock()
	defer self.Unlock()
	return self.aliveCt
}
func (self *Node) decAlive() {
	self.Lock()
	defer self.Unlock()
	self.aliveCt--
}

func (self *Node) init() bool {
	self.connChan = make(chan net.Conn, 128)
	self.timerConnCheck()
	return true

}

func (self *Node) timerConnCheck() {
	self.initAllConn()
	tk := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-tk.C:
				self.initAllConn()
			}
		}
	}()
}

func (self *Node) initAllConn() {
	num := 12 - len(self.connChan)
	for i := 0; i < num; i++ {
		conn, err := net.Dial("tcp", self.Host)
		if err != nil {
			log.Printf("net dial host fail: %v", self.Host)
			return
		}
		self.connChan <- conn
		self.incAlive()
		log.Printf("connect to %s succ", self.Host)
	}
}

func (self *Node) Dowork(pbp *PBDataPack) *PBDataPack {
	if pbp == nil {
		return nil
	}
	conn := <-self.connChan
	if pbp.Send(conn) == false {
		self.decAlive()
		return nil
	}
	ret := &PBDataPack{}
	if ret.Unpack(conn) == false {
		self.decAlive()
		log.Printf("conn lost %s connchan len: %d", self.Host, len(self.connChan))
		return nil
	}
	log.Printf("send to %s connchan len: %d", self.Host, len(self.connChan))
	self.connChan <- conn
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
		if node.GetAlive() == 0 {
			continue
		}
		self._totalWeight += node.Weight
	}

	//calc startw & endw
	startW := uint32(0)
	for _, node := range self._lst {
		if node.GetAlive() == 0 {
			continue
		}
		node.StartW = startW
		node.EndW = startW + node.Weight
		startW = node.EndW
	}
	log.Printf("after dorebal totalWeight: %d, alive node count: %d", self._totalWeight, self._aliveCount())
}

func (self *NodeList) init() {
	rand.Seed(time.Now().UnixNano())

	//born all node from cfg
	cfg := GetCfgData("")
	self._totalWeight = uint32(0)
	for _, elem := range cfg.Ary {
		self._totalWeight += elem.Weight
	}

	startW := uint32(0)
	for _, elem := range cfg.Ary {
		node := &Node{
			aliveCt: 0,
		}
		node.Host = elem.Host
		node.StartW = startW
		node.EndW = startW + elem.Weight
		node.Weight = elem.Weight
		self._lst = append(self._lst, node)
		startW = node.EndW
	}

	log.Printf("begin connect to all hostconnect to all host")
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
		if nd.GetAlive() != 0 {
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
		if nd.GetAlive() == 0 {
			bfind = true
		}
	}
	if bfind {
		self._doRebalance()
	}
}

func (self *NodeList) Dispatch(pbp *PBDataPack) (*PBDataPack, int) {
	var _nd *Node
	{
		self.Lock()
		x := rand.Intn(int(self._totalWeight))
		log.Printf("Dispatch total: %d ,nowrand: %d", self._totalWeight, x)

		for _, nd := range self._lst {
			if nd.GetAlive() == 0 {
				continue
			}
			if uint32(x) >= nd.StartW && uint32(x) < nd.EndW {
				_nd = nd
				break
			}
		}
		self.Unlock()
	}

	ret := _nd.Dowork(pbp)
	if ret != nil {
		return ret, 0
	}

	{
		self.Lock()
		//not get at all
		if self._aliveCount() > 0 {
			self._doRebalance()
			//need redo dispatch
			self.Unlock()
			return nil, 1
		}
		self.Unlock()
	}
	log.Printf("Dispatch fail, no alived server")
	return nil, 0
}
