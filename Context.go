package proc

// Context xxx
type Context interface {
	Self() ProcessId
	Message() (interface{}, ProcessId)
}
