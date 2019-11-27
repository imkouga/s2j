package v2

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/imkouga/s2j"
)

// Marshal return a string.
func Marshal(objects interface{}, auth s2j.AuthType) (v interface{}, err error) {
	authMap, err := buildAuth(auth)
	if err != nil {
		return nil, err
	}

	values := reflect.ValueOf(objects)
	if values.Kind() == reflect.Ptr {
		values = values.Elem()
	}

	log.SetOutput(os.Stdout)

	switch values.Kind() {
	case reflect.Slice, reflect.Array:
		var wg sync.WaitGroup
		nums := values.Len()
		vs := make([]map[string]interface{}, nums, nums)
		wg.Add(nums)
		for i := 0; i < nums; i++ {
			go func(i int) {
				defer wg.Done()
				s2m, err := m(values.Index(i), authMap, "")
				if err != nil {
					log.Printf("数据鉴权出错。错误原因:%s", err.Error())

					return
				}
				// vs = append(vs, s2m)
				if s2m != nil && len(s2m) != 0 {
					vs[i] = s2m
				}
			}(i)
		}
		wg.Wait()

		return vs, nil

	case reflect.Struct:
		s2m, err := m(values, authMap, "")
		return s2m, err

	default:
		msg := fmt.Sprintf("一级数据类型必须是数组或者切片或者结构体类型, 类型 id为%d", reflect.TypeOf(objects).Kind())
		return nil, s2j.InvalidObjects{Msg: msg}
	}
}

func m(object reflect.Value, auth map[string]bool, preTag string) (v map[string]interface{}, err error) {
	if object.Kind() == reflect.Ptr {
		object = object.Elem()
	}

	switch object.Kind() {
	case reflect.Struct:
		v = make(map[string]interface{})
		num := object.NumField()
		t := object.Type()
		for i := 0; i < num; i++ {
			tag := t.Field(i).Tag.Get("json")
			if len(tag) == 0 {
				return nil, s2j.InvalidObjects{Msg: "struct tag must be required."}
			}

			tagKey := strings.TrimLeft(fmt.Sprintf("%s.%s", preTag, tag), ".")
			field := object.Field(i)
			if field.Kind() == reflect.Ptr {
				field = field.Elem()
			}

			switch field.Kind() {
			case reflect.Array, reflect.Slice:
				childLen := field.Len()
				vv := make([]map[string]interface{}, childLen, childLen)
				isNull := true
				for ii := 0; ii < childLen; ii++ {
					s2m, err := m(field.Index(ii), auth, tagKey)
					if err != nil {
						return nil, err
					}
					if s2m != nil && len(s2m) != 0 {
						log.Printf("not nil %v", s2m)
						isNull = false
						vv[ii] = s2m
					}
				}
				if !isNull {
					v[tag] = vv
				}

			case reflect.Bool:
				fallthrough
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fallthrough
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fallthrough
			case reflect.Float32, reflect.Float64:
				fallthrough
			case reflect.String:
				if _, found := auth[tagKey]; found && auth[tagKey] {
					v[tag] = field.Interface()
				}

			case reflect.Struct:
				switch field.Interface().(type) {
				case time.Time, *time.Time:
					if _, found := auth[tagKey]; found && auth[tagKey] {
						v[tag] = field.Interface()
					}

				default:
					s2m, err := m(field, auth, tagKey)
					if err != nil {
						return nil, err
					}
					if s2m != nil && len(s2m) != 0 {
						v[tag] = s2m
					}
				}

			default:
				if _, found := auth[tagKey]; found && auth[tagKey] {
					v[tag] = nil
				}
			}
		}

	default:
		return nil, s2j.InvalidObjects{Msg: "objects must be struct"}

	}

	return v, nil
}

// 深度搜索算法
// type test11Auth struct {
//	A bool `json:"a"`
//	B bool `json:"b"`
// }
// type test1Auth struct {
//	A bool        `json:"a"`
//	B bool        `json:"b"`
//	C *test11Auth `json:"c"`
//	D *test11Auth `json:"d"`
//}
//构建完得出
// map["a"] = true
// map["b"] = true
// map["c.a"] = true
// map["c.b"] = true
// map["d.a"] = true
// map["d.b"] = true
func buildAuth(auth s2j.AuthType) (map[string]bool, error) {
	authMap := make(map[string]bool)
	value := reflect.ValueOf(auth)
	err := dfsBuildAuth(authMap, "", value)

	return authMap, err
}

func dfsBuildAuth(authMap map[string]bool, curTag string, value reflect.Value) (err error) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Bool && value.Kind() != reflect.Struct {
		return s2j.InvalidAuthType{Msg: fmt.Sprintf("无效的Auth 对象，其字段类型必须是布尔类型或者结构体类型 - %d", value.Kind())}
	}

	if value.Kind() == reflect.Bool {
		var authBool, ok bool

		auth := value.Interface()
		if authBool, ok = auth.(bool); !ok {
			return s2j.InvalidAuthType{Msg: "无效的Auth对象，其字段值必须是布尔类型"}
		}
		authMap[curTag] = authBool

		return nil
	}

	nums := value.NumField()
	t := value.Type()
	for i := 0; i < nums; i++ {
		tag := t.Field(i).Tag.Get("json")
		if len(tag) == 0 {
			return s2j.InvalidAuthType{Msg: "无效的Auth对象，其结构体的Tag标签必须提供json标签"}
		}

		dfsBuildAuth(authMap, strings.TrimLeft(fmt.Sprintf("%s.%s", curTag, tag), "."), value.Field(i))
	}

	return nil
}
