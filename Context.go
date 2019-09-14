package proc

// Context xxx
type Context interface {
	Message() (interface{}, ProcessID)
}
