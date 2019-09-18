package proc

import "time"

// Self xxx
type Self = func() ProcessID

// Receive xxx
type Receive = func(ProcessID, interface{})

// ReceiveDispatcher xxxx
type ReceiveDispatcher = func(Receive, ...time.Duration)

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(Self, ReceiveDispatcher, ...interface{})

// Spawn xxxx
func Spawn(handler ProcessHandler, args ...interface{}) ProcessID {

	return newProcess(handler, args).
		start()
}
