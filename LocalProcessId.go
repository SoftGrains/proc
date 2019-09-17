package proc

type localProcessID struct {
	process *process
}

func (pid *localProcessID) Send(message interface{}) {

	pid.process.Send(message)
}

func (pid *localProcessID) SendFrom(sender ProcessID, message interface{}) {

	pid.process.Send(messageWithSender{
		sender:  sender,
		message: message,
	})
}
