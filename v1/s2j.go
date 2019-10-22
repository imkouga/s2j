package v1

import (
	"reflect"
	"sync"

	"github.com/imkouga/s2j"
)

// Marshal return a string.
func Marshal(objects interface{}, auth interface{}) (v []map[string]interface{}, err error) {
	var (
		wg sync.WaitGroup
		vm []map[string]interface{}
		ok bool
		l  sync.Mutex
	)

	vm = make([]map[string]interface{}, 0)

	if _, ok = auth.(s2j.AuthType); !ok {
		return nil, s2j.InvalidAuthType{}
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
				if object.Kind() == reflect.Ptr {
					object = object.Elem()
				}
				objectRef := object.Type()
				_o := make(map[string]interface{})
				for i := 0; i < object.NumField(); i++ {
					if _, found := authMap[objectRef.Field(i).Name]; found {
						if authMap[objectRef.Field(i).Name] {
							_o[objectRef.Field(i).Tag.Get("json")] = object.Field(i).Interface()
						}
					}
				}
				l.Lock()
				vm = append(vm, _o)
				l.Unlock()
			}(object)
		}
		wg.Wait()
	default:
		return nil, s2j.InvalidObjects{Msg: "objects must be a array or slice"}
	}

	return vm, nil
}
