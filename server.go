package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func startServer() {
	l, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		panic("dang")
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic("pang")
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	retChan := make(chan *arg)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	go recvLoop(reader, retChan)
	go sendLoop(writer, retChan)
}

func recvLoop(reader *bufio.Reader, retChan chan *arg) {
	for {
		// 1. read method
		method, err := readMethod(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading: ", err.Error())
			}
			fmt.Println("Client half closed")
			break
		}

		// 2. handle method
		fmt.Printf("server parsed method: %v\n", method)
		ret := call(method)
		retChan <- ret
	}

}

func sendLoop(writer *bufio.Writer, retChan chan *arg) {
	for {
		arg := <-retChan
		buf := arg.serialize()
		writeLen := 0

		for writeLen < len(buf) {
			len, err := writer.Write(buf[writeLen:])
			if err != nil {
				panic("wat write failed")
			}
			writeLen += len
		}
		writer.Flush()
	}
}

func call(method *method) *arg {
	// do some silly matching here
	if method.name == "hello" {
		return &arg{0, []byte(hello())}
	} else if method.name == "add" {
		args := method.args
		a := int(binary.BigEndian.Uint32(args[0].val))
		b := int(binary.BigEndian.Uint32(args[1].val))
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(add(a, b)))
		return &arg{1, buf}
	}
	return &arg{}
}

func readMethod(reader *bufio.Reader) (*method, error) {
	totalLenBuf, err := readNBytes(reader, 4)
	if err != nil {
		return nil, err
	}

	totalLen := int(binary.BigEndian.Uint32(totalLenBuf))
	methodBuf, err := readNBytes(reader, totalLen)
	if err != nil {
		return nil, err
	}
	buf := append(totalLenBuf, methodBuf...)
	method, _ := deserializeMethod(buf, 0)
	return method, err
}

func readNBytes(reader io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	readLen := 0
	for readLen < n {
		len, err := reader.Read(buf[readLen:])
		if err != nil {
			return nil, err
		}
		readLen += len
	}
	return buf, nil
}

func hello() string {
	return "hello world from ssun"
}

func add(a int, b int) int {
	return a + b
}
