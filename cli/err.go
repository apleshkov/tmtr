package cli

type BadFlagErr struct {
	msg string
}

func newBadFlag(msg string) *BadFlagErr {
	return &BadFlagErr{msg}
}

func (e *BadFlagErr) Error() string {
	return e.msg
}
