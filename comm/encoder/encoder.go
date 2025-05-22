//lint:file-ignore ST1006 receiver name should be a reflection of its identity
package encoder

import (
	"encoding/json"
)

type JSONEncoder interface {
	EncodeToByte(obj any) ([]byte, error)

	EncodeToString(obj any) (*string, error)

	DecodeFromString(jsonStr *string, obj any) error

	DecodeFromBytes(jsonBytes []byte, obj any) error
}

type GoEncoder struct{}

func (self *GoEncoder) EncodeToByte(obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func (self *GoEncoder) EncodeToString(obj any) (*string, error) {
	jsonByte, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	jsonStr := string(jsonByte)
	return &jsonStr, nil
}

func (self *GoEncoder) DecodeFromString(jsonStr *string, obj any) error {
	return json.Unmarshal([]byte(*jsonStr), &obj)
}

func (self *GoEncoder) DecodeFromBytes(jsonByte []byte, obj any) error {
	return json.Unmarshal(jsonByte, &obj)
}
