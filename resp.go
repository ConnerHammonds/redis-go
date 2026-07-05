// Deserializer for RESP requests
package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// first time using a struct in go.
// They are similar to python classes, but
// they don't perform any operations on the data.
// Good for things that have an is-a or has-a relationship
// structs can be passed into functions to create a similar
// effect that python methods have.
type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	// return line minus last two bytes.
	// minus 2 because the last two bytes are the crlf line endings (\r\n)
	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	// initialze new Value struct and assign typ to "array"
	v := Value{}
	v.typ = "array"

	// Read next byte in the resp request which is the length of the array
	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// For each line, parse and read the value
	v.array = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		// add parsed value to array
		v.array[i] = val
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	// new value struct and assign typ to "bulk"
	v := Value{}
	v.typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)

	r.reader.Read(bulk)
	fmt.Println("Hello World")

	v.bulk = string(bulk)

	// Read trailing crlf
	// Without r.readLine(), the pointer would be stuck
	// at \r\n from the previous string and wouldn't be
	// able to read the next string in the array
	r.readLine()

	return v, nil
}
