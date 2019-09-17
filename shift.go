package bl3_auto_vip

type ShiftConfig struct {
	CodeListUrl string `json:"codeListUrl"`
	CodeListRowSelector string `json:"codeListRowSelector"`
	CodeListInvalidRegex string `json:"codeListInvalidRegex"`
	CodeListCheckIndex int `json:"codeListCheckIndex"`
	CodeListCodeIndex int `json:"codeListCodeIndex"`
	SessionHeader string `json:"sessionHeader"`
}