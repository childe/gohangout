package simplejson

/**
MOST CODE IS COPYED FROM GOLANG JSON PACKAGE!
since gohangout event has limited data type(int/float/map/array/time/string, NO struct), so we can use a simple way to do json encode
**/

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unicode/utf8"
)

type SimpleJsonDecoder struct {
	bytes.Buffer
	scratch [64]byte
}

type JSONMarshaler interface {
	MarshalJSON() ([]byte, error)
}

var hex = "0123456789abcdef"

func (d *SimpleJsonDecoder) string(s string) int {
	len0 := d.Len()
	d.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' && b != '<' && b != '>' && b != '&' {
				i++
				continue
			}
			if start < i {
				d.WriteString(s[start:i])
			}
			switch b {
			case '\\', '"':
				d.WriteByte('\\')
				d.WriteByte(b)
			case '\n':
				d.WriteByte('\\')
				d.WriteByte('n')
			case '\r':
				d.WriteByte('\\')
				d.WriteByte('r')
			case '\t':
				d.WriteByte('\\')
				d.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				d.WriteString(`\u00`)
				d.WriteByte(hex[b>>4])
				d.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				d.WriteString(s[start:i])
			}
			d.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				d.WriteString(s[start:i])
			}
			d.WriteString(`\u202`)
			d.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		d.WriteString(s[start:])
	}
	d.WriteByte('"')
	return d.Len() - len0
}

func (d *SimpleJsonDecoder) encodeV(v interface{}) error {
	if v == nil {
		d.WriteString("null")
		return nil
	}

	k := reflect.TypeOf(v).Kind()
	switch k {
	case reflect.Bool:
		if v.(bool) {
			d.WriteString("true")
		} else {
			d.WriteString("false")
		}

	case reflect.Int:
		b := strconv.AppendInt(d.scratch[:0], int64(v.(int)), 10)
		d.Write(b)
	case reflect.Int32:
		b := strconv.AppendInt(d.scratch[:0], int64(v.(int32)), 10)
		d.Write(b)
	case reflect.Int64:
		b := strconv.AppendInt(d.scratch[:0], v.(int64), 10)
		d.Write(b)

	case reflect.Float32:
		d.encodeFloat(float64(v.(float32)), 32)
	case reflect.Float64:
		d.encodeFloat(v.(float64), 32)
	case reflect.String:
		// it could be either string or json.Number
		d.string(reflect.ValueOf(v).String())
	case reflect.Map:
		return d.encodeMap(v.(map[string]interface{}))
	case reflect.Slice, reflect.Array:
		return d.encodeSlice(v)
	default:
		if o, ok := v.(JSONMarshaler); ok {
			if b, err := o.MarshalJSON(); err != nil {
				return err
			} else {
				d.Write(b)
				return nil
			}
		}
		return fmt.Errorf("unknownType %T", v)
	}
	return nil
}

func (d *SimpleJsonDecoder) encodeSlice(value interface{}) error {
        switch value.(type) {
	case []byte:
                d.string(string(value.([]byte)))
	default:
	        t := reflect.ValueOf(value)           	
                d.WriteByte('[')
	        n := t.Len()
	        for i := 0; i < n; i++ {
                        if i > 0 {
	                        d.WriteByte(',')
                        }
	                d.encodeV(t.Index(i).Interface())
	       }
	       d.WriteByte(']')
	       return nil
        }
	return nil
}

func (d *SimpleJsonDecoder) encodeFloat(f float64, bits int) error {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return fmt.Errorf("UnsupportedValueError %v %s", f, strconv.FormatFloat(f, 'g', -1, int(bits)))
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}
	b := strconv.AppendFloat(d.scratch[:0], f, fmt, -1, bits)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	d.Write(b)
	return nil
}

func (d *SimpleJsonDecoder) encodeMap(e map[string]interface{}) error {
	if e == nil {
		d.WriteString("null")
		return nil
	}

	d.WriteByte('{')

	var i = 0
	for k, v := range e {
		if i > 0 {
			d.WriteByte(',')
		}
		d.WriteByte('"')
		_, err := d.WriteString(k)
		if err != nil {
			return err
		}
		d.WriteByte('"')
		d.WriteByte(':')

		err = d.encodeV(v)
		if err != nil {
			return err
		}
		i++
	}
	d.WriteByte('}')
	return nil
}

func (d *SimpleJsonDecoder) encodeArray(v []interface{}) error {
	d.WriteByte('[')
	n := len(v)
	for i := 0; i < n; i++ {
		if i > 0 {
			d.WriteByte(',')
		}
		d.encodeV(v[i])
	}
	d.WriteByte(']')
	return nil
}

func (d *SimpleJsonDecoder) Encode(e interface{}) ([]byte, error) {
	if err := d.encodeV(e); err != nil {
		return nil, err
	}
	return d.Bytes(), nil
}
