package server

type EOFError struct {
	err error
}

func (m EOFError) Error() string {
	return m.err.Error()
}
