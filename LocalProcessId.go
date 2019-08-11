package proc

// ProcessId xxx
type localProcessId struct {
	process *process
}

func (pid *localProcessId) Send(message interface{}) {

	pid.process.Send(
		message,
		false)
}

func (pid *localProcessId) SendFrom(sender ProcessId, message interface{}) {

	pid.Send(
		messageWithSender{
			sender:  sender,
			message: message,
		})	
}


func (pid *localProcessId) Stop() {

	pid.process.Send(
		stopMessage{}, true)
}
