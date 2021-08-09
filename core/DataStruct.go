package core

import (
	"net/http"
)

type Options struct {
	Thread     int
	Timeout    int
	Headers    []HTTPHeader
	Dictionary []string
	DirRoot    string
	Target     []string
	Cookie     string
	Method     string
}

type ReqRes struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Length     int64
}

type WildCard struct {
	StatusCode int
	Location   string
	Body       []byte
	Length     int64
	Type       int
}

type WafCk struct {
	WafName string
	Alive   bool
}
