package proc

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(ProcessID, ...interface{}) func(ProcessID, interface{})

// Spawn xxxx
func Spawn(handler ProcessHandler, args ...interface{}) ProcessID {

	return newProcess(handler, args).
		start()
}
