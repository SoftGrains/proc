package proc

// ProcessId xxx
type ProcessId interface {
	Send(message interface{})
	SendFrom(sender ProcessId, message interface{})
	Stop()
}
