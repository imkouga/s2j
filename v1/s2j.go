package v1

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/imkouga/s2j"
)

// Marshal return a string.
func Marshal(objects interface{}, auth interface{}) (v string, err error) {
	var (
		vb []byte
		wg sync.WaitGroup
		vm []map[string]interface{}
		ok bool
	)

	vm = make([]map[string]interface{}, 0)

	if _, ok = auth.(s2j.AuthType); !ok {
		return "", s2j.InvalidAuthType{}
	}

	switch reflect.TypeOf(objects).Kind() {
	case reflect.Array, reflect.Slice:
		authMap, _ := s2j.ParseAuth(auth)
		objectsOfValue := reflect.ValueOf(objects)

		for index := 0; index < objectsOfValue.Len(); index++ {
			object := objectsOfValue.Index(index)
			wg.Add(1)
			go func(object reflect.Value) {
				defer wg.Done()
				objectRef := object.Type()
				_o := make(map[string]interface{})
				for i := 0; i < object.NumField(); i++ {
					if _, found := authMap[objectRef.Field(i).Name]; found {
						if authMap[objectRef.Field(i).Name] {
							_o[objectRef.Field(i).Tag.Get("json")] = object.Field(i).Interface()
						}
					} else {
						_o[objectRef.Field(i).Tag.Get("json")] = object.Field(i).Interface()
					}
				}
				vm = append(vm, _o)
			}(object)
		}
		wg.Wait()
	default:
		return "", s2j.InvalidObjects{Msg: "objects must be a array or slice"}
	}

	if vb, err = json.Marshal(vm); err != nil {
		return "", err
	}

	return string(vb), nil
}