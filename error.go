package s2j

type InvalidAuthType struct{}

func (InvalidAuthType) Error() string {
	return "invalid auth type."
}

type InvalidObjects struct {
	Msg string
}

func (i InvalidObjects) Error() string {
	return i.Msg
}
