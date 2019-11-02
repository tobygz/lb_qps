package main

import (
	//"bytes"
	//"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
    _ "net/http/pprof"
    "log"
    "net/http"

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

	var l net.Listener
	var err error
	l, err = net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + ":" + *port)

    if *mode == 1 {
        GetNodeList().init(*cfg)
    }

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting", err.Error())
			return // 终止程序
		}
		if *mode == 1 {
			go handReqProxy(conn)
		} else if *mode == 2 {
			go handReqServ(conn)
		} else {
			//route
			go handReqRawProxy(conn)
		}
	}

}

//for server
func handReqServ(conn net.Conn) {
	fmt.Println("accept conn handReqServ")
	pp := &PBDataPack{}

	for {
		//read from client
		if !pp.Unpack(conn) {
			break
		}
		//send back
		if !pp.Send(conn) {
			break
		}
	}
}

//for proxy
func handReqRawProxy(conn net.Conn) {
	fmt.Println("accept conn handReqRawProxy")
	remote_conn, err := net.Dial("tcp", *remote)
	if err != nil {
		panic(err)
	}
	go func() {
		//io.Copy(remote_conn,conn)
		bt := make([]byte, 1)
		for {
			if _, err := io.ReadFull(remote_conn, bt); err != nil {
				fmt.Println(err)
				break
			}
			if _, err := conn.Write(bt); err != nil {
				fmt.Println(err)
				break
			}
			fmt.Print(int(bt[0]), "_")
		}
	}()
	for {
		//io.Copy(conn,remote_conn)
		bt := make([]byte, 1)
		if _, err := io.ReadFull(conn, bt); err != nil {
			fmt.Println(err)
			break
		}
		if _, err := remote_conn.Write(bt); err != nil {
			fmt.Println(err)
			break
		}
		fmt.Print(int(bt[0]), "+")
	}
}

//for proxy
func handReqProxy(conn net.Conn) {
	fmt.Println("accept conn handReqProxy")
	pp := &PBDataPack{}
	/*
		ppr := &PBDataPack{}
		remote_conn, err := net.Dial("tcp", *remote)
		if err != nil {
			panic(err)
		}
	*/
	for {
		//read from client
		if !pp.Unpack(conn) {
            continue
		}
		ppr := GetNodeList().Dispatch(pp)
		if ppr == nil {
			fmt.Println("dispatch fail, is all server crashed")
            continue
		}
		/*
			//send to remote
			if !pp.Send(remote_conn) {
				break
			}
			//read from remote
			if !ppr.Unpack(remote_conn) {
				break
			}
		*/
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
