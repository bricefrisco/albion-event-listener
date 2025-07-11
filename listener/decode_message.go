package listener

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	nilType        = 42
	dictionaryType = 68
	int8Type       = 98
	float32Type    = 102
	int32Type      = 105
	int16Type      = 107
	int64Type      = 108
	booleanType    = 111
	stringType     = 115
	int8SliceType  = 120
	sliceType      = 121
)

func decodeReliableMessage(msg reliableMessage) (map[uint8]any, error) {
	buf := bytes.NewBuffer(msg.data)
	params := make(map[uint8]interface{})

	for i := 0; i < int(msg.parameterCount); i++ {
		var paramId uint8
		var paramType uint8

		if err := binary.Read(buf, binary.BigEndian, &paramId); err != nil {
			return nil, err
		}

		if err := binary.Read(buf, binary.BigEndian, &paramType); err != nil {
			return nil, err
		}

		paramValue, err := decodeType(buf, paramType)
		if err != nil {
			return nil, err
		}

		params[paramId] = paramValue
	}

	return params, nil
}

func decodeType(buf *bytes.Buffer, paramType uint8) (any, error) {
	switch paramType {
	case nilType:
		return nil, nil
	case dictionaryType:
		return decodeDictionary(buf)
	case int8Type:
		return decodeInt8(buf)
	case float32Type:
		return decodeFloat32(buf)
	case int32Type:
		return decodeInt32(buf)
	case int16Type:
		return decodeInt16(buf)
	case int64Type:
		return decodeInt64(buf)
	case booleanType:
		return decodeBoolean(buf)
	case stringType:
		return decodeString(buf)
	case int8SliceType:
		return decodeInt8Slice(buf)
	case sliceType:
		return decodeSliceType(buf)
	default:
		return nil, fmt.Errorf("unsupported type %d", paramType)
	}
}

func decodeSliceType(buf *bytes.Buffer) (any, error) {
	var length uint16
	var sType uint8

	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.BigEndian, &sType); err != nil {
		return nil, err
	}

	switch sType {
	case int8Type:
		return decodeInt8Slice(buf)
	case int16Type:
		return decodeInt16Slice(buf, length)
	case int32Type:
		return decodeInt32Slice(buf, length)
	case int64Type:
		return decodeInt64Slice(buf, length)
	case float32Type:
		return decodeFloat32Slice(buf, length)
	case stringType:
		return decodeStringSlice(buf, length)
	case booleanType:
		return decodeBooleanSlice(buf, length)
	case sliceType:
		return decodeSliceSlice(buf, length)
	default:
		return nil, fmt.Errorf("unsupported type %d", sType)
	}
}

func decodeInt8(buf *bytes.Buffer) (value int8, err error) {
	err = binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeInt16(buf *bytes.Buffer) (value int16, err error) {
	err = binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeInt32(buf *bytes.Buffer) (value int32, err error) {
	err = binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeInt64(buf *bytes.Buffer) (value int64, err error) {
	err = binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeFloat32(buf *bytes.Buffer) (value float32, err error) {
	err = binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeString(buf *bytes.Buffer) (string, error) {
	var length uint16
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return "", err
	}

	str := make([]byte, length)
	if _, err := buf.Read(str); err != nil {
		return "", err
	}

	return string(str), nil
}

func decodeBoolean(buf *bytes.Buffer) (bool, error) {
	var value uint8
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return false, err
	}

	if value == 0 {
		return false, nil
	} else if value == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid value for boolean of %d", value)
	}
}

func decodeDictionary(buf *bytes.Buffer) (map[interface{}]interface{}, error) {
	var keyTypeCode uint8
	var valueTypeCode uint8
	var length uint16

	if err := binary.Read(buf, binary.BigEndian, &keyTypeCode); err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.BigEndian, &valueTypeCode); err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	dictionary := make(map[interface{}]interface{}, length)
	for i := 0; i < int(length); i++ {
		key, err := decodeType(buf, keyTypeCode)
		if err != nil {
			return nil, err
		}

		value, err := decodeType(buf, valueTypeCode)
		if err != nil {
			return nil, err
		}

		dictionary[key] = value
	}

	return dictionary, nil
}

func decodeInt8Slice(buf *bytes.Buffer) (any, error) {
	var length uint32

	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	byteSlice := make([]byte, length)
	if _, err := buf.Read(byteSlice); err != nil {
		return nil, err
	}

	int8Slice := make([]int8, length)
	for i, b := range byteSlice {
		int8Slice[i] = int8(b)
	}

	return int8Slice, nil
}

func decodeInt16Slice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]int16, length)
	err := binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeInt32Slice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]int32, length)
	err := binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeInt64Slice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]int64, length)
	err := binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeFloat32Slice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]float32, length)
	err := binary.Read(buf, binary.BigEndian, &value)
	return value, err
}

func decodeStringSlice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]string, length)
	for i := 0; i < int(length); i++ {
		str, err := decodeString(buf)
		if err != nil {
			return nil, err
		}
		value[i] = str
	}
	return value, nil
}

func decodeBooleanSlice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]bool, length)
	for i := 0; i < int(length); i++ {
		b, err := decodeBoolean(buf)
		if err != nil {
			return nil, err
		}
		value[i] = b
	}
	return value, nil
}

func decodeSliceSlice(buf *bytes.Buffer, length uint16) (any, error) {
	value := make([]interface{}, length)
	for i := 0; i < int(length); i++ {
		subArray, err := decodeSliceType(buf)
		if err != nil {
			return nil, err
		}
		value[i] = subArray
	}
	return value, nil
}
