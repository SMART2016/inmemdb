package server

import (
	"errors"
)

/**
- https://redis.io/docs/reference/protocol-spec/#:~:text=RESP%20can%20serialize%20different%20data,that%20the%20server%20should%20execute
- Request Response Protocol (RESP serialization / Deserialization Protocol
   - It is light weight as compared to JSON
   - Commands are sent by redis client as an array of strings serialised using RESP.
   - Every data type starts with a special character
   - Data ends with \r\n (CRLF)
     - Eg:
       - Sending a **string** pong to redis : `+pong\r\n`
       - Sending an **int**: `:1720\r\n`
       - **Bulk Strings**: `$4\r\npong\r\n`
         - Starts with $
         - 4 is the number of bytes in the string.
         - Bulk strings are binary safe, simple strings cannot contain \r\n in the string itself as they are the terminators for the end of string
         - With bulk string we can store any binary data in redis.
         - Null value: `$-1\r\n` (-1 tells no data)
       - **Array**: `["a",200,"cat]`
         - Starts with *
         - RESP encoding of above array:
         - `*3\r\n
            $1\r\na\r\n
            :200\r\n
            $3\r\ncat\r\n`
         - empty array : `*0\r\n`
         - null array : `*-1\r\n`
       - **Error**:
         - Starts with -
         - `- Key Not Found \r\n`
*/

func Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	value, _, err := decodeFirstElement(data)
	return value, err
}

func decodeFirstElement(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}
	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}
	return nil, 0, nil
}

// reads the length typically the first integer of the string
// until hit by an non-digit byte and returns
// the integer and the delta = length + 2 (CRLF)
// TODO: Make it simpler and read until we get `\r` just like other functions
func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

// reads a RESP encoded simple string from data and returns
// the string, the delta, and the error
func readSimpleString(data []byte) (string, int, error) {
	// first character +
	pos := 1

	for ; data[pos] != '\r'; pos++ {
	}

	return string(data[1:pos]), pos + 2, nil
}

// reads a RESP encoded error from data and returns
// the error string, the delta, and the error
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// reads a RESP encoded integer from data and returns
// the intger value, the delta, and the error
func readInt64(data []byte) (int64, int, error) {
	// first character :
	pos := 1
	var value int64 = 0

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}

	return value, pos + 2, nil
}

// reads a RESP encoded string from data and returns
// the string, the delta, and the error
func readBulkString(data []byte) (string, int, error) {
	// first character $
	pos := 1

	// reading the length and forwarding the pos by
	// the lenth of the integer + the first special character
	len, delta := readLength(data[pos:])
	pos += delta

	// reading `len` bytes as string
	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

// reads a RESP encoded array from data and returns
// the array, the delta, and the error
func readArray(data []byte) (interface{}, int, error) {
	// first character *
	pos := 1

	// reading the length
	count, delta := readLength(data[pos:])
	pos += delta

	var elems []interface{} = make([]interface{}, count)
	for i := range elems {
		elem, delta, err := decodeFirstElement(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elems[i] = elem
		pos += delta
	}
	return elems, pos, nil
}

/**
All commands are sent to the server as an array of strings encoded as RESP.
This function helps decode the input command into a pure array of strings from the RESP encoded array of strings
	Input Command: *3\r\n$3\r\nPUT\r\n$1\r\nK\r\n$1\r\nV\r\n
	Decoded Output : ["PUT","K","V"]
*/
func DecodeInputCommand(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	ts := value.([]interface{})
	tokens := make([]string, len(ts))
	for i := range tokens {
		tokens[i] = ts[i].(string)
	}

	return tokens, nil
}
