package proc

// Context xxx
type Context interface {
	Self() ProcessID
	Message() (interface{}, ProcessID)
}
