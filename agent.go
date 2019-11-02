package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var port = flag.String("port", "3333", "port")
var remote = flag.String("remote", "127.0.0.1:443", "")
var cfg = flag.String("cfg", "s.json", "")
var mode = flag.Int("mode", 0, "0: client 1: agent, 2: server")

func main() {
	flag.Parse()

	//test_cfg()
	//return
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	if *mode == 0 {
		//for client
		do_client()
		return
	}

	if *mode == 1 {
		GetNodeList().init(*cfg)
	}

	var l net.Listener
	var err error
	l, err = net.Listen("tcp", ":"+g_lbElemAry.Port)
	if err != nil {
		log.Println("Error listening: %v", err)
		os.Exit(1)
	}
	defer l.Close()
	log.Println("Listening on " + ":" + *port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: %v", err.Error())
			return
		}
		go handReqProxy(conn)
	}

}

//for proxy
func handReqProxy(conn net.Conn) {
	log.Println("accept conn handReqProxy")
	pp := &PBDataPack{}
	for {
		//read from client
		if !pp.Unpack(conn) {
			continue
		}
		ppr := GetNodeList().Dispatch(pp)
		if ppr == nil {
			log.Println("dispatch fail, is all server crashed")
			continue
		}
		//send to client
		if !ppr.Send(conn) {
			continue
		}
	}
}

//// 1. 12-byte header [PRPC][body_size][meta_size]
func do_client() {
	remote_conn, err := net.Dial("tcp", *remote)
	if err != nil {
		panic(err)
	}

	ppr := &PBDataPack{}
	ppr.Head.Name = make([]byte, 4)
	ppr.Head.Name = []byte("2345")
	bodysize := uint32(12)
	metasize := uint32(14)
	ppr.Head.Bodysize = bodysize
	ppr.Head.Metasize = metasize
	ppr.Body = make([]byte, bodysize)
	ppr.Send(remote_conn)

	ppget := &PBDataPack{}
	ppget.Unpack(remote_conn)
	ppget.Tostring()
}
