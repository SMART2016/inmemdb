package server

/**
COmmands sent to the server are generally an operation followed by a set of arguments
Example: PUT Key Value
*/
type Command struct {
	//The operation to perform
	Cmd string

	//The Arguments to perform the operation
	Args []string
}
