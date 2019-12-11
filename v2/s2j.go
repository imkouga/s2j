package v2

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/imkouga/s2j"
)

//type S2j struct {
//	valueBigMap map[string]string //一个大的 map ，记录需要数据鉴权的数据字段的 tag
//	valueBigMapLock sync.Mutex
//}
//
//func (tj *S2j)setValueBigMap(key, value string) {
//	tj.valueBigMapLock.Lock()
//	tj.valueBigMap[key] = value
//	tj.valueBigMapLock.Unlock()
//}
//
//func GetValueBigMap(key string) (value string, ok bool) {
//	tj.valueBigMapLock.Lock()
//	if _, ok = valueBigMap[key]; !ok {
//		tj.valueBigMapLock.Unlock()
//		return "", ok
//	}
//
//	value = valueBigMap[key]
//	valueBigMapLock.Unlock()
//
//	return value, true
//}

// Marshal return a big map.
func Marshal(objects interface{}, auth s2j.AuthType) (v interface{}, err error) {
	authMap, authTagMap, err := buildAuth(auth)
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
		var l sync.Mutex
		nums := values.Len()
		vs := make([]map[string]interface{}, 0, nums)
		wg.Add(nums)
		for i := 0; i < nums; i++ {
			go func(i int) {
				defer wg.Done()
				s2m, err := m(values.Index(i), authMap, authTagMap, "")
				if err != nil {
					log.Printf("数据鉴权出错。错误原因:%s", err.Error())

					return
				}
				if s2m != nil && len(s2m) != 0 {
					l.Lock()
					vs = append(vs, s2m)
					l.Unlock()
				}
			}(i)
		}
		wg.Wait()

		return vs, nil

	case reflect.Struct:
		s2m, err := m(values, authMap, authTagMap, "")
		return s2m, err

	default:
		msg := fmt.Sprintf("一级数据类型必须是数组或者切片或者结构体类型, 类型 id为%d", reflect.TypeOf(objects).Kind())
		return nil, s2j.InvalidObjects{Msg: msg}
	}
}

func m(object reflect.Value, auth map[string]bool, authTag map[string]string, preName string) (v map[string]interface{}, err error) {
	if object.Kind() == reflect.Ptr {
		object = object.Elem()
	}

	switch object.Kind() {
	case reflect.Struct:
		v = make(map[string]interface{})
		num := object.NumField()
		t := object.Type()
		var (
			buf bytes.Buffer
		)
		buf.Grow(40)
		for i := 0; i < num; i++ {
			buf.Reset()
			if len(preName) != 0 {
				buf.WriteString(preName)
				buf.WriteString(".")
			}
			buf.WriteString(t.Field(i).Name)
			curPathName := buf.String()

			field := object.Field(i)
			if field.Kind() == reflect.Ptr {
				field = field.Elem()
			}

			switch field.Kind() {
			case reflect.Array, reflect.Slice:
				childLen := field.Len()
				vv := make([]map[string]interface{}, 0, childLen)
				isNull := true
				var (
					wgii   sync.WaitGroup
					wglock sync.Mutex
				)
				wgii.Add(childLen)
				for ii := 0; ii < childLen; ii++ {
					go func(ii int) {
						defer wgii.Done()
						s2m, err := m(field.Index(ii), auth, authTag, curPathName)
						if err != nil {
							log.Printf("数据权限执行有误。%s", err.Error())
						}
						if s2m != nil && len(s2m) != 0 {
							isNull = false
							wglock.Lock()
							vv = append(vv, s2m)
							wglock.Unlock()
						}
					}(ii)
				}
				wgii.Wait()

				if !isNull {
					if _, found := authTag[curPathName]; found {
						v[authTag[curPathName]] = vv
					}
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
				if _, found := auth[curPathName]; found && auth[curPathName] {
					if _, found := authTag[curPathName]; found {
						v[authTag[curPathName]] = field.Interface()
					}
				}

			case reflect.Struct:
				switch field.Interface().(type) {
				case time.Time, *time.Time:
					if _, found := auth[curPathName]; found && auth[curPathName] {
						if _, found := authTag[curPathName]; found {
							v[authTag[curPathName]] = field.Interface()
						}
					}

				default:
					s2m, err := m(field, auth, authTag, curPathName)
					if err != nil {
						return nil, err
					}
					if s2m != nil && len(s2m) != 0 {
						if _, found := authTag[curPathName]; found {
							v[authTag[curPathName]] = s2m
						}
					}
				}

			default:
				if _, found := auth[curPathName]; found && auth[curPathName] {
					if _, found := authTag[curPathName]; found {
						v[authTag[curPathName]] = nil
					}
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
// map["A"] = true
// map["B"] = true
// map["C.A"] = true
// map["C.B"] = true
// map["D.A"] = true
// map["D.B"] = true

// map["A"] = a
// map["B"] = b
// map["C.A"] = a
// map["C.B"] = b
// map["D.A"] = a
// map["D.B"] = b
func buildAuth(auth s2j.AuthType) (map[string]bool, map[string]string, error) {
	authMap := make(map[string]bool)
	authTagMap := make(map[string]string)
	value := reflect.ValueOf(auth)
	err := dfsBuildAuth(authMap, authTagMap, "", "", value)

	return authMap, authTagMap, err
}

func dfsBuildAuth(authMap map[string]bool, authTagMap map[string]string, namePath string, curTag string, value reflect.Value) (err error) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Bool && value.Kind() != reflect.Struct {
		return s2j.InvalidAuthType{Msg: fmt.Sprintf("无效的Auth 对象，其字段类型必须是布尔类型或者结构体类型 - %d", value.Kind())}
	}

	if len(namePath) != 0 {
		authTagMap[namePath] = curTag
	}

	if value.Kind() == reflect.Bool {
		var authBool, ok bool

		auth := value.Interface()
		if authBool, ok = auth.(bool); !ok {
			return s2j.InvalidAuthType{Msg: "无效的Auth对象，其字段值必须是布尔类型"}
		}
		authMap[namePath] = authBool

		return nil
	}

	nums := value.NumField()
	t := value.Type()
	var buf bytes.Buffer
	buf.Grow(40)
	for i := 0; i < nums; i++ {
		tag := t.Field(i).Tag.Get("json")
		if len(tag) == 0 {
			return s2j.InvalidAuthType{Msg: "无效的Auth对象，其结构体的Tag标签必须提供json标签"}
		}

		tagIndex := strings.Index(tag, ",")
		if tagIndex == -1 {
			tagIndex = len(tag)
		}

		tag = tag[0:tagIndex]

		if len(namePath) != 0 {
			buf.WriteString(namePath)
			buf.WriteString(".")
		}
		buf.WriteString(t.Field(i).Name)
		dfsBuildAuth(authMap, authTagMap, buf.String(), tag, value.Field(i))
		buf.Reset()
	}

	return nil
}
