package controller

import (
	mycode "github.com/mou-he/graduation-design/common/code"
)

// Response 响应体
type Response struct {
	StatusCode mycode.Code `json:"status_code"`
	StatusMsg  string      `json:"status_msg,omitempty"`
}

func (r *Response) CodeOf(code mycode.Code) Response {
	if nil == r {
		r = new(Response)
	}
	r.StatusCode = code
	r.StatusMsg = code.Msg()
	return *r
}

func (r *Response) Success() {
	r.CodeOf(mycode.CodeSuccess)
}
