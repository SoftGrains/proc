package proc

type messageWithSender struct {
	sender  ProcessID
	message interface{}
}

type localProcessID struct {
	process *process
}

func (pid *localProcessID) Send(message interface{}) {

	var _, internal = message.(StopProcessMessage)

	pid.process.Send(
		message,
		internal)
}

func (pid *localProcessID) SendFrom(sender ProcessID, message interface{}) {

	var _, internal = message.(StopProcessMessage)

	pid.process.Send(
		messageWithSender{
			sender:  sender,
			message: message,
		},
		internal)
}
