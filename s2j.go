package s2j

import (
	"reflect"
)

type AuthType interface {
	AuthName() string
}

type S2J interface {
	ModelName() string
}

func ParseAuth(auth interface{}) (map[string]bool, error) {
	authOfValue := reflect.ValueOf(auth)
	authOfType := reflect.TypeOf(auth)
	authMap := make(map[string]bool)
	for i := 0; i < authOfValue.NumField(); i++ {
		authMap[authOfType.Field(i).Name] = authOfValue.Field(i).Bool()
	}

	return authMap, nil
}
