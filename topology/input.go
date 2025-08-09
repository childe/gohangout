package topology

type Input interface {
	ReadOneEvent() map[string]any
	Shutdown()
}
