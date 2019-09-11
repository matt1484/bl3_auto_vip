package bl3_auto_vip

import (
	"github.com/thedevsaddam/gojsonq"
)

type StringSet map[string]struct{}

func (set StringSet) Add(s string) {
	set[s] = struct{}{}
}

func JsonFromString(s string) *gojsonq.JSONQ {
	return gojsonq.New().JSONString(s)
}

func JsonFromBytes(bytes []byte) *gojsonq.JSONQ {
	return JsonFromString(string(bytes))
}
