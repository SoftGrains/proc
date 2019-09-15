package proc

// ProcessID xxx
type ProcessID interface {
	Send(message interface{})
	SendFrom(sender ProcessID, message interface{})
}
