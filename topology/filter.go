package topology

type Filter interface {
	Filter(map[string]interface{}) (map[string]interface{}, bool)
}
