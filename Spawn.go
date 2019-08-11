package proc

//process handler
type Actor = func(...interface{}) func(Context)

func Spawn(fn Actor, args ...interface{}) ProcessId {

	var proc = newProcess(fn, args)
	proc.start()

	return proc.ID()
}
