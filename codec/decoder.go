package codec

type Decoder interface {
	Decode([]byte) map[string]interface{}
}

func NewDecoder(t string) Decoder {
	switch t {
	case "plain":
		return &PlainDecoder{}
	case "json":
		return &JsonDecoder{useNumber: true}
	case "json:not_usenumber":
		return &JsonDecoder{useNumber: false}
	}
	panic(t + " decoder not supported")
	return nil
}
