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

type Bl3Config struct {
	Version string `json:"version"`
	LoginUrl string `json:"loginUrl"`
	LoginRedirectHeader string `json:"loginRedirectHeader"`
	SessionIdHeader string `json:"sessionIdHeader"`
	RequestHeaders map[string]string `json:"requestHeaders"`
	SessionHeader string `json:"sessionHeader"`
	Vip VipConfig `json:"vipConfig"`
	Shift ShiftConfig `json:"shiftConfig"`
}