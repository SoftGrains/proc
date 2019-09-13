package proc

// Actor xxxxxxxxxx
type Actor = func(...interface{}) func(Context)

// Spawn xxxx
func Spawn(fn Actor, args ...interface{}) ProcessID {

	var proc = newProcess(fn, args)
	proc.start()

	return proc.ID()
}
