package proc

type processContext struct {
	message interface{}
}

func newContext(message interface{}) Context {

	return &processContext{
		message: message,
	}
}

// Message xxx
func (context *processContext) Message() (interface{}, ProcessID) {

	switch msg := context.message.(type) {

	case messageWithSender:
		return msg.message, msg.sender
	}

	return context.message, nil
}
