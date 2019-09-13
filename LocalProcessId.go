package proc

type messageWithSender struct {
	sender  ProcessID
	message interface{}
}

type localProcessID struct {
	process *process
}

func (pid *localProcessID) Send(message interface{}) {

	pid.process.Send(
		message,
		false)
}

func (pid *localProcessID) SendFrom(sender ProcessID, message interface{}) {

	pid.Send(
		messageWithSender{
			sender:  sender,
			message: message,
		})
}

func (pid *localProcessID) Stop() {

	pid.process.Send(
		stopProcessMessage{}, true)
}
