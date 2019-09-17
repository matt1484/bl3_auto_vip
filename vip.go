package bl3_auto_vip

import (
	// "fmt"
	"errors"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
)

type VipCodeMap map[string]StringSet

func (v VipCodeMap) Diff(other VipCodeMap) VipCodeMap {
	diff := VipCodeMap{}
	for codeType, codes := range v {
		for code := range codes {
			if _, found := other[codeType][code]; !found {
				if _, found := diff[codeType]; !found {
					diff[codeType] = StringSet{}
				}
				diff[codeType].Add(code)
			}
		}
	}
	return diff
}

func (v VipCodeMap) Add(codeType, code string) {
	codeType = strings.ToLower(codeType)
	code = strings.ToLower(code)
	if _, found := v[codeType]; found {
		v[codeType].Add(code)
	}
}

type VipConfig struct {
	CodeListUrl string `json:"codeListUrl"`
	CodeListRowSelector string `json:"codeListRowSelector"`
	CodeListInvalidRegex string `json:"codeListInvalidRegex"`
	CodeListCheckIndex int `json:"codeListCheckIndex"`
	CodeListCodeIndex int `json:"codeListCodeIndex"`
	CodeListTypeIndex int `json:"codeListTypeIndex"`
	CodeTypeUrlMap map[string]string  `json:"codeTypeUrlMap"`
}

func (conf *VipConfig) GetCodeTypes() []string {
	codeTypes := make([]string, 0)
	for codeType, _ := range conf.CodeTypeUrlMap {
		codeTypes = append(codeTypes, codeType)
	}
	return codeTypes
}

func (conf *VipConfig) DetectCodeTypes(s string) []string {
	l := strings.ToLower(s)
	types := make([]string, 0)
	for codeType, _ := range conf.CodeTypeUrlMap {
		if strings.Contains(l, codeType) {
			types = append(types, codeType)
		}
	}
	return types
}

func (conf *Bl3Config) NewVipCodeMap() VipCodeMap {
	codeTypeMap := make(map[string]StringSet)
	for codeType, _ := range conf.Vip.CodeTypeUrlMap {
		codeTypeMap[codeType] = StringSet{}
	}
	return codeTypeMap
}

func (client *Bl3Client) GetFullVipCodeMap() (VipCodeMap, error) {
	codeMap := client.Config.NewVipCodeMap()
	httpClient, err := NewHttpClient()
	if err != nil {
		return codeMap, err
	}

	response, err := httpClient.Get("https://www.reddit.com/r/borderlands3/comments/bxgq5p/borderlands_vip_program_codes/")
	if err != nil {
		return codeMap, errors.New("Failed to get code list")
	}

	redditHtml, err := response.BodyAsHtmlDoc()
	if err != nil {
		return codeMap, err
	}

	redditHtml.Find("[data-test-id='post-content'] tbody tr").Each(func(i int, row *goquery.Selection) {
		if len(row.Find("td").Nodes) < 4 {
			return
		}

		codeTypes := ""
		code := ""

		row.Find("td").EachWithBreak(func(i int, col *goquery.Selection) bool {
			if i == 2 && strings.Contains(strings.ToLower(col.Text()), "no") {
				return false
			}
			if i == 0 {
				code = strings.ToLower(col.Text())
			}
			if i == 3 {
				codeTypes = strings.ToLower(col.Text())
				return false
			}
			return true
		})

		for _, codeType := range client.Config.Vip.DetectCodeTypes(codeTypes) {
			codeMap[codeType].Add(code)
		}
	})
	return codeMap, nil
}

func (client *Bl3Client) GetRedeemedVipCodeMap() (VipCodeMap, error) {
	codeMap := client.Config.NewVipCodeMap()

	url := "https://2kgames.crowdtwist.com/request?widgetId=9470"
	data := map[string]interface{}{
		"model_data": map[string]interface{}{
			"activity": map[string]interface{}{
				"newest_activities": map[string]interface{}{
					"properties": []string{"notes", "title"},
					"query": map[string]interface{}{
						"type": "user_activities_me",
						"args": map[string]int{
							"row_start": 1,
							"row_end":   1000000,
						},
					},
				},
			},
		},
	}

	res, err := client.PostJson(url, data)
	if err != nil {
		return codeMap, errors.New("Failed to get redeemed code list")
	}

	type activity struct {
		CodeTypes string `json:"title"`
		Code      string `json:"notes"`
	}

	activities := make([]activity, 0)
	resJson, err := res.BodyAsJson()
	if err != nil {
		return codeMap, err
	}

	resJson.From("model_data.activity.newest_activities").Out(&activities)

	for _, act := range activities {
		for _, codeType := range client.Config.Vip.DetectCodeTypes(act.CodeTypes) {
			codeMap.Add(codeType, act.Code)
		}
	}
	return codeMap, nil
}

func (client *Bl3Client) getVipWidgetConf(url string) *gojsonq.JSONQ {
	response, err := client.Get(url)
	if err != nil {
		return nil
	}

	widgetHtml, err := response.BodyAsHtmlDoc()
	if err != nil {
		return nil
	}

	script := ""

	widgetHtml.Find("script").EachWithBreak(func(i int, scriptTag *goquery.Selection) bool {
		if strings.Contains(scriptTag.Text(), "widgetConf") {
			script = scriptTag.Text()
			return false
		}
		return true
	})

	script = strings.TrimSpace(strings.Split(strings.Join(strings.Split(strings.Join(strings.Split(script, "widgetConf")[1:], ""), "=")[1:], ""), ";")[0])
	json := JsonFromString(script)
	return json
}

func (client *Bl3Client) GenerateVipCodeUrlMap() (map[string]string, error) {
	codeTypeUrlMap := make(map[string]string)

	widgetConf := client.getVipWidgetConf("https://2kgames.crowdtwist.com/widgets/t/activity-list/9904/?__locale__=en#2")
	if widgetConf == nil {
		return codeTypeUrlMap, errors.New("Failed to get code redemption types")
	}

	type widget struct {
		WidgetId   int    `json:"widgetId"`
		WidgetName string `json:"widgetName"`
	}

	widgets := make([]widget, 0)
	widgetConf.From("entries").Select("link.widgetId", "link.widgetName").Out(&widgets)

	for _, wid := range widgets {
		for _, codeType := range client.Config.Vip.DetectCodeTypes(wid.WidgetName) {
			widgetConf := client.getVipWidgetConf("https://2kgames.crowdtwist.com/widgets/t/code-redemption/" + strconv.Itoa(wid.WidgetId))
			if widgetConf == nil {
				codeTypeUrlMap[codeType] = ""
				continue
			}

			campaignId, ok := widgetConf.Find("campaignId").(float64)
			if !ok {
				codeTypeUrlMap[codeType] = ""
				continue
			}
			codeTypeUrlMap[codeType] = "https://2kgames.crowdtwist.com/code-redemption-campaign/redeem?cid=" + strconv.Itoa(int(campaignId))
		}
	}
	
	return codeTypeUrlMap, nil
}

func (client *Bl3Client) RedeemVipCode(codeType, code string) (string, bool) {
	res, err := client.PostJson(client.Config.Vip.CodeTypeUrlMap[codeType], map[string]string {
		"code": code,
	})
	if err != nil {
		return "bad request", false
	}

	resJson, err := res.BodyAsJson()
	if err != nil {
		return "invalid response", false
	}

	exception := ""
	resJson.From("exception.model").Out(&exception)
	success := ""
	resJson.Reset().From("message").Out(&success)
	if exception != "" {
		// technically the code may be valid but just unredeemable by this account (limits/already redeemed)
		return exception, !strings.Contains(strings.ToLower(exception), "invalid")
	}
	if success != "" {
		return success, true
	}

	return "wrong response format", false
}