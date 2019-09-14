package proc

// ProcessHandler xxxxxxxxxx
type ProcessHandler = func(ProcessID, ...interface{}) func(Context)

// Spawn xxxx
func Spawn(fn ProcessHandler, args ...interface{}) ProcessID {

	var proc = newProcess(fn, args)
	proc.start()

	return proc.ID()
}
