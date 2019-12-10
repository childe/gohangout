package topology

type Output interface {
	Emit(map[string]interface{})
	Shutdown()
}
