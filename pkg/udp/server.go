package udp

type ConnectionStatus int

const (
	Joining ConnectionStatus = iota
	Leaving
)
