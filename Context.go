package proc

import "time"

// Context xxx
type Context struct {
	self    ProcessID
	receive ReceiveFunction
	args    []interface{}
}

func newContext(self ProcessID,
	receive ReceiveFunction,
	args []interface{}) *Context {

	return &Context{
		self:    self,
		receive: receive,
		args:    args,
	}
}

// Self xxxx
func (context *Context) Self() ProcessID {

	return context.self
}

// Receive xxxx
func (context *Context) Receive(receive ReceiveHandler) {

	context.receive(receive)
}

// ReceiveTimeout xxxx
func (context *Context) ReceiveTimeout(after time.Duration, receive ReceiveHandler) {

	context.receive(receive, after)
}

// Arguments xxx
func (context *Context) Arguments() []interface{} {

	return context.args
}
