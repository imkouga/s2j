package v1_test

import (
	"testing"

	v1 "github.com/imkouga/s2j/v1"
)

type lll struct {
	Aa string `json:"aa"`
	Bb string `json:"bb"`
}
type lllAuth struct {
	Aa bool
	Bb bool
}

func (lllAuth) AuthName() string {
	return ""
}

func TestAuth(t *testing.T) {
	ll := []lll{lll{Aa: "454", Bb: "444"}}
	llAuth := lllAuth{Aa: true, Bb: true}

	_, err := v1.Marshal(ll, llAuth)
	if err != nil {
		t.Fatalf("marshal failed. err is %s", err.Error())
	}
}
