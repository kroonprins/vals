package base64

import (
	"encoding/base64"
)

type base64Encode struct {
}
type base64Decode struct {
}

func NewBase64Encode() *base64Encode {
	p := &base64Encode{}
	return p
}
func NewBase64Decode() *base64Decode {
	p := &base64Decode{}
	return p
}

func (p *base64Encode) Convert(value string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(value)), nil
}

func (p *base64Decode) Convert(value string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", nil
	}
	return string(decoded), nil
}
