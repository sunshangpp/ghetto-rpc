package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {

	go startServer()
	time.Sleep(100 * time.Millisecond)

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:9999")

	fmt.Println("client call hello...")
	hello := &method{"hello", nil}
	buf := hello.serialize()
	writeCompletely(conn, buf)
	ret, _ := readArg(conn)
	retVal := ret.getValue()
	fmt.Printf("client return: %v\n", retVal)

	time.Sleep(100 * time.Millisecond)

	fmt.Println("client call add...")
	args := make([]*arg, 2)
	args[0] = &arg{1, intToBytes(10)}
	args[1] = &arg{1, intToBytes(99)}
	add := &method{"add", args}
	buf = add.serialize()
	writeCompletely(conn, buf)
	ret, _ = readArg(conn)
	retVal = ret.getValue()
	fmt.Printf("client return: %v\n", retVal)

}

func intToBytes(n int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	return buf
}

func writeCompletely(writer io.Writer, buf []byte) {
	writeLen := 0
	for writeLen < len(buf) {
		len, err := writer.Write(buf[writeLen:])
		if err != nil {
			panic("write failed from client")
		}
		writeLen += len
	}
}

func readArg(reader io.Reader) (*arg, error) {
	totalLenBuf, err := readNBytes(reader, 4)
	if err != nil {
		return nil, err
	}

	totalLen := int(binary.BigEndian.Uint32(totalLenBuf))
	argBuf, err := readNBytes(reader, totalLen)
	if err != nil {
		return nil, err
	}
	buf := append(totalLenBuf, argBuf...)
	arg, _ := deserializeArg(buf, 0)
	return arg, err
}
