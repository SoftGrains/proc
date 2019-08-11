package proc

type processContext struct {
	self    ProcessId
	message interface{}
}

// NewContext xxxx
func newContext(self ProcessId, message interface{}) Context {

	return &processContext{
		self:    self,
		message: message,
	}
}

// Self xxx
func (context *processContext) Self() ProcessId {
	return context.self
}

// Message xxx
func (context *processContext) Message() (interface{}, ProcessId) {

	switch msg := context.message.(type) {

	case messageWithSender:
		return msg.message, msg.sender
	}

	return context.message, nil
}

type messageWithSender struct {
	sender  ProcessId
	message interface{}
}
