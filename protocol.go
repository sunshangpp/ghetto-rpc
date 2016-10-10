package main

import "encoding/binary"

/*
Method:
4 bytes: method name len
n bytes: method name
4 bytes: number of arguments
n bytes: serialized arguments

Arg:
4 bytes: content length (including type)
1 byte : argument type
n bytes: argument value

value type table:
+-------------+-----------+
|    TYPE     |    CODE   |
+-------------+-----------+
|   string    |   0       |
|   int       |   1       |
+-------------+-----------+

*/

type method struct {
	name string
	args []*arg
}

// used for method argument and return value
type arg struct {
	typ int
	val []byte
}

func (arg *arg) getValue() interface{} {
	if arg.typ == 0 {
		return string(arg.val)
	} else if arg.typ == 1 {
		return int(binary.BigEndian.Uint32(arg.val))
	}
	return nil
}

func (method *method) serialize() []byte {
	buf := make([]byte, 4)
	name := method.name
	binary.BigEndian.PutUint32(buf, uint32(len(name)))
	buf = append(buf, []byte(name)...)

	buf = append(buf, make([]byte, 4)...)
	binary.BigEndian.PutUint32(buf[4+len(name):], uint32(len(method.args)))

	for _, arg := range method.args {
		buf = append(buf, arg.serialize()...)
	}

	// first 4 bytes for content length
	buf = append(make([]byte, 4), buf...)
	binary.BigEndian.PutUint32(buf, uint32(len(buf)-4))

	return buf
}

func (arg *arg) serialize() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(1+len(arg.val)))
	buf = append(buf, byte(arg.typ))
	buf = append(buf, arg.val...)
	return buf
}

func deserializeMethod(buf []byte, offset int) (*method, int) {
	// skip first 4 bytes of total content length, it is only used for socket read
	offset += 4

	methodNameLen := int(binary.BigEndian.Uint32(buf[offset : offset+4]))
	offset += 4

	methodName := string(buf[offset : offset+methodNameLen])
	offset += methodNameLen

	numArgs := int(binary.BigEndian.Uint32(buf[offset : offset+4]))
	offset += 4

	args := make([]*arg, numArgs)
	for i := 0; i < numArgs; i++ {
		var arg *arg
		arg, offset = deserializeArg(buf, offset)
		args[i] = arg
	}

	return &method{
		name: methodName,
		args: args,
	}, offset
}

func deserializeArg(buf []byte, offset int) (*arg, int) {
	len := int(binary.BigEndian.Uint32(buf[offset : offset+4]))
	return &arg{
		typ: int(buf[offset+4]),
		val: buf[offset+5 : offset+4+len],
	}, offset + 4 + len
}
