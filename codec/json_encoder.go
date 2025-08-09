package codec

type JsonEncoder struct{}

func (e *JsonEncoder) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}
