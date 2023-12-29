package response

import "fmt"

/**
This function is used to Encode the response from the server in RESP form before sending back to the client.
	- Currently we have added all response to be encoded as string or binary string (Bulk string)
*/
func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}
	return []byte{}
}

func EncodeError(err error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", err))
}
