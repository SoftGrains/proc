package proc

// ReceiveHandler xxxx
type ReceiveHandler = func(func(ProcessID, interface{}))

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(ProcessID, ReceiveHandler, ...interface{})

// Spawn xxxx
func Spawn(handler ProcessHandler, args ...interface{}) ProcessID {

	return newProcess(handler, args).
		start()
}
