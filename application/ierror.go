package main

type ierror struct {
	e error
	m string
}

func (e ierror) Unwrap() error {
	return e.e
}

func (e ierror) Error() string {
	if e.e != nil {
		return e.e.Error()
	}
	return ""
}

func (e ierror) Message() string {
	return e.m
}
