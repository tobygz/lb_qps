package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type PkgHead struct {
	Name     []byte
	Bodysize uint32
	Metasize uint32
}

func (this *PkgHead) Unpack(bin []byte) bool {
	// 读取Len
	this.Name = make([]byte, 4)
	this.Name = bin[:4]
	headbuf := bytes.NewReader(bin[4:])

	if err := binary.Read(headbuf, binary.BigEndian, &this.Bodysize); err != nil {
		fmt.Println(err)
		return false
	}
	if err := binary.Read(headbuf, binary.BigEndian, &this.Metasize); err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("Unpack name: %s bodysize: %d metasize: %d", this.Name, this.Bodysize, this.Metasize)
	return true
}

func (this *PkgHead) Pack() []byte {
	binW := bytes.NewBuffer([]byte{})
	binW.WriteString(string(this.Name))

	binary.Write(binW, binary.BigEndian, this.Bodysize)
	binary.Write(binW, binary.BigEndian, this.Metasize)
	fmt.Println("pack name: %s bodysize: %d metasize: %d", this.Name, this.Bodysize, this.Metasize)
	return binW.Bytes()
}

func (this *PkgHead) ToString() {
	fmt.Println(this.Name, this.Bodysize, this.Metasize)
}

type PBDataPack struct {
	Head PkgHead
	Body []byte
}

func (this *PBDataPack) Unpack(conn net.Conn) bool {

	headbin := make([]byte, 12)

	if _, err := io.ReadFull(conn, headbin); err != nil {
		fmt.Println(err)
		return false
	}

	this.Head.Unpack(headbin)
	//this.Body = make([]byte, this.Head.Bodysize+this.Head.Metasize)
	this.Body = make([]byte, this.Head.Bodysize)
	if _, err := io.ReadFull(conn, this.Body); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (this *PBDataPack) Tostring() {
	fmt.Println(string(this.Head.Name), this.Head.Bodysize, this.Head.Metasize, len(this.Body))
}

func (this *PBDataPack) Send(conn net.Conn) bool {
	hbin := this.Head.Pack()
	if _, err := conn.Write(hbin); err != nil {
		return false
	}
	if _, err := conn.Write(this.Body); err != nil {
		return false
	}
	return true
}

func PkgHead___test() {
	t0 := &PkgHead{}
	t0.Name = []byte("prbc")
	t0.Bodysize = uint32(32)
	t0.Metasize = uint32(63)

	bin := t0.Pack()
	t0.ToString()

	fmt.Println("size:", len(bin))
	t1 := &PkgHead{}
	t1.Unpack(bin)
	t1.ToString()
}