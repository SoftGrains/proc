package proc

// ReceiveContext xxx
type ReceiveContext struct {
	sender  ProcessID
	message interface{}
}

func newReceiveContext(sender ProcessID,
	message interface{}) *ReceiveContext {

	return &ReceiveContext{
		sender:  sender,
		message: message,
	}
}

// Sender xxxx
func (context *ReceiveContext) Sender() ProcessID {

	return context.sender
}

// Message xxxx
func (context *ReceiveContext) Message() interface{} {

	return context.message
}
