package codec

import (
	"errors"

	"github.com/childe/gohangout/value_render"
)

type FormatEncoder struct {
	render value_render.ValueRender
}

var ErrNotString = errors.New("value returned by FormatEncoder is not a string type")

func NewFormatEncoder(format string) *FormatEncoder {
	return &FormatEncoder{
		render: value_render.GetValueRender(format),
	}
}

func (e *FormatEncoder) Encode(v any) ([]byte, error) {
	rst, err := e.render.Render(v.(map[string]any))
	if err != nil {
		return nil, err
	}

	if v, ok := rst.(string); ok {
		return []byte(v), nil
	}
	return nil, ErrNotString
}
