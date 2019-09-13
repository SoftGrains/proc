package proc

type processContext struct {
	pid     ProcessID
	message interface{}
}

func newContext(pid ProcessID, message interface{}) Context {

	return &processContext{
		pid:     pid,
		message: message,
	}
}

// Self xxx
func (context *processContext) Self() ProcessID {
	return context.pid
}

// Message xxx
func (context *processContext) Message() (interface{}, ProcessID) {

	switch msg := context.message.(type) {

	case messageWithSender:
		return msg.message, msg.sender
	}

	return context.message, nil
}
