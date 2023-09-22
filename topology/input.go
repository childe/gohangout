package topology

type Input interface {
	ReadOneEvent() map[string]interface{}
	Pause()
	Shutdown()
}
