package proc

import "time"

func timerProcess(context *Context) {

	var args = context.Arguments()

	if len(args) < 2 {
		return
	}

	sender, ok := args[0].(ProcessID)
	if ok == false ||
		sender == nil {
		return
	}

	timeoutAfter, ok := args[1].(time.Duration)
	if ok == false ||
		timeoutAfter < 0 {
		return
	}

	// sleep for a timeout
	time.Sleep(timeoutAfter)

	sender.SendFrom(context.Self(), timeoutMessage{
		After: timeoutAfter,
	})
}
