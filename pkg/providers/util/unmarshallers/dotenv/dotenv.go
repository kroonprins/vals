package dotenv

import (
	"github.com/joho/godotenv"
)

type unmarshaller struct{}

func New() *unmarshaller {
	return &unmarshaller{}
}

func (*unmarshaller) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	res, err := godotenv.Unmarshal(string(bytes))
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	for k, v := range res {
		m[k] = v
	}
	return m, nil
}
