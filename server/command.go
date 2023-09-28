package server

import (
	"errors"
	"log"
)

/**
Commands sent to the server are generally an operation followed by a set of arguments
Example: PUT Key Value
*/
type Command struct {
	//The operation to perform
	Cmd string

	//The Arguments to perform the operation
	Args []string
}

func (cmd *Command) EvalCommand() ([]byte, error) {
	log.Println("comamnd:", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return cmd.evalPING()
	default:
		return cmd.evalPING()
	}
}

func (cmd *Command) evalPING() ([]byte, error) {
	var b []byte

	if len(cmd.Args) >= 2 {
		return nil, errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(cmd.Args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(cmd.Args[0], false)
	}

	return b, nil
}
