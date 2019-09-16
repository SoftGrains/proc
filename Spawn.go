package proc

// Receive xxx
type Receive = func(ProcessID, interface{})

// ReceiveDispatcher xxxx
type ReceiveDispatcher = func(Receive)

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(ProcessID, ReceiveDispatcher, ...interface{})

// Spawn xxxx
func Spawn(handler ProcessHandler, args ...interface{}) ProcessID {

	return newProcess(handler, args).
		start()
}
