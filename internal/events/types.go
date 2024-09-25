package events

type Consumer interface {
	Start() error
	StartV1() error
}
