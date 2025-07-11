package listener

import (
	"fmt"
	"strconv"
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func sanitizeValues(v any) any {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			m[fmt.Sprintf("%v", k)] = sanitizeValues(v2)
		}
		return m

	case map[byte]any:
		m := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			m[strconv.Itoa(int(k))] = sanitizeValues(v2)
		}
		return m

	case map[string]interface{}:
		for k, v2 := range val {
			val[k] = sanitizeValues(v2)
		}
		return val

	case []interface{}:
		for i := range val {
			val[i] = sanitizeValues(val[i])
		}
		return val

	case []uint8: // treat as bytes, convert to list of ints
		ints := make([]int, len(val))
		for i, b := range val {
			ints[i] = int(b)
		}
		return ints

	default:
		return val
	}
}

func toMessage(msgType string, msg map[uint8]any) *Message {
	vals := sanitizeValues(msg)

	m := &Message{
		Type: msgType,
		Data: vals,
	}

	return m
}
