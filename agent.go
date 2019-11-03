package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
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

	GetNodeList().init(*cfg)
	timerLog()

	log.Printf("port:%v", g_lbElemAry.Port)
	var l net.Listener
	var err error
	l, err = net.Listen("tcp", ":"+g_lbElemAry.Port)
	if err != nil {
		log.Printf("Error listening: %v", err)
		os.Exit(1)
	}
	defer l.Close()
	log.Printf("Listening on " + ":" + g_lbElemAry.Port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting: %v", err.Error())
			return
		}
		go handReqProxy(conn)
	}

}

//for proxy
func handReqProxy(conn net.Conn) {
	log.Printf("accept conn handReqProxy")
	pp := &PBDataPack{}
	ch := make(chan *PBDataPack, 1024)
	for i := 0; i < 12; i++ {
		go func() {
			for {
				select {
				case tpp := <-ch:
					ppr, tp := GetNodeList().Dispatch(tpp)
					if ppr == nil {
						if tp == 1 {
							ch <- tpp
							continue
						}
						log.Printf("dispatch fail, is all server crashed")
						continue
					}
					//send to client
					if !ppr.Send(conn) {
						break
					}
				}
			}
		}()
	}
	for {
		//read from client
		if !pp.Unpack(conn) {
			break
		}
		ch <- pp
		log.Printf("chan len: %d", len(ch))
	}
}
func rotateLog() {
	t := time.Unix(time.Now().Unix(), 0)
	nt := t.Format("2006_01_02_15")

	fname := fmt.Sprintf("%s/%s_%s.log", g_lbElemAry.LogDir, "/agent", nt)
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("fatal error ", err)
		return
	}
	log.SetOutput(f)
}

func timerLog() {
	rotateLog()
	tk := time.NewTicker(time.Hour)
	go func() {
		for {
			select {
			case <-tk.C:
				rotateLog()
			}
		}
	}()
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
