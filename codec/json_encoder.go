package codec

type JsonEncoder struct{}

func (e *JsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
