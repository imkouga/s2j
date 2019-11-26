package v2

import (
	"encoding/json"
	"testing"
)

type test11 struct {
	A int64  `json:"a"`
	B string `json:"b"`
}
type test11Auth struct {
	A bool `json:"a"`
	B bool `json:"b"`
}

func (test11Auth) AuthName() string {
	return ""
}

type test1 struct {
	A int64   `json:"a"`
	B string  `json:"b"`
	C *test11 `json:"c"`
	D test11  `json:"d"`
}
type test1Auth struct {
	A bool        `json:"a"`
	B bool        `json:"b"`
	C *test11Auth `json:"c"`
	D *test11Auth `json:"d"`
	E test11Auth  `json:"e"`
}

func (test1Auth) AuthName() string {
	return ""
}

func TestAuth(t *testing.T) {

	auth := &test1Auth{
		A: true,
		B: false,
		C: &test11Auth{
			A: false,
			B: true,
		},
		D: &test11Auth{
			A: true,
			B: false,
		},
		E: test11Auth{
			A: true,
			B: true,
		},
	}

	data := &test1{
		A: 1,
		B: "dfs",
		C: &test11{
			A: 3,
			B: "dfsg",
		},
		D: test11{
			A: 5,
			B: "ggg",
		},
	}

	datas := make([]*test1, 0, 1)
	datas = append(datas, data)

	v, err := Marshal(datas, auth)
	if err != nil {
		t.Fatal(err)
	}

	str, _ := json.Marshal(v)
	t.Logf("%s", string(str))
}

func TestBuildAuth(t *testing.T) {
	auth := &test1Auth{
		A: false,
		B: true,
		C: &test11Auth{
			A: true,
			B: false,
		},
		D: &test11Auth{
			A: true,
			B: true,
		},
		E: test11Auth{
			A: true,
			B: true,
		},
	}

	authMap, err := buildAuth(auth)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", authMap)
}
