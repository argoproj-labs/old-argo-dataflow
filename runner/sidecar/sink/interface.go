package sink

type Interface interface {
	Sink(msg []byte) error
}
