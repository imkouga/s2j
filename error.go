package s2j

type InvalidAuthType struct {
	Msg string
}

func (i InvalidAuthType) Error() string {
	if i.Msg == "" {
		return "invalid auth type."
	}

	return i.Msg
}

type InvalidObjects struct {
	Msg string
}

func (i InvalidObjects) Error() string {
	return i.Msg
}
