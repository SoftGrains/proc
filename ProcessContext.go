package proc

type processContext struct {
	self    ProcessID
	message interface{}
}

// NewContext xxxx
func newContext(self ProcessID, message interface{}) Context {

	return &processContext{
		self:    self,
		message: message,
	}
}

// Self xxx
func (context *processContext) Self() ProcessID {
	return context.self
}

// Message xxx
func (context *processContext) Message() (interface{}, ProcessID) {

	switch msg := context.message.(type) {

	case messageWithSender:
		return msg.message, msg.sender
	}

	return context.message, nil
}

type messageWithSender struct {
	sender  ProcessID
	message interface{}
}
