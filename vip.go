package bl3_auto_vip

import (
	"errors"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
	"strconv"
)

type VipCodeTypeMap map[string]string

type VipCodeMap map[string]StringSet

func (v VipCodeMap) Diff(other VipCodeMap) VipCodeMap {
	diff := NewVipCodeMap()
	for codeType, codes := range v {
		for code, _ := range codes {
			if _, found := other[codeType][code]; !found {
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

func GetCodeTypes() []string {
	return []string {
		"pax",
		"boost",
		"email",
		"creator",
		"vault",
		"diamond",
	}
}

func GetCodeTypesInString(s string) []string {
	l := strings.ToLower(s)
	types := make([]string, 0)
	for _, t := range GetCodeTypes() {
		if strings.Contains(l, t) {
			types = append(types, t)
		}
	}
	return types
}

func NewVipCodeTypeMap() VipCodeTypeMap {
	codeTypeMap := make(map[string]string)
	for _, codeType := range GetCodeTypes() {
		codeTypeMap[codeType] = ""
	}
	return codeTypeMap
}

func NewVipCodeMap() VipCodeMap {
	codeTypeMap := make(map[string]StringSet)
	for _, codeType := range GetCodeTypes() {
		codeTypeMap[codeType] = StringSet{}
	}
	return codeTypeMap
}

type Bl3VipClient struct {
	HttpClient
}

func NewBl3VipClient() (*Bl3VipClient, error) {
	client, err := NewHttpClient()
	if err != nil {
		return nil, errors.New("Failed to start client")
	}

	client.SetDefaultHeader("Origin", "https://borderlands.com")
	client.SetDefaultHeader("Referer", "https://borderlands.com/en-US/vip/")

	return &Bl3VipClient{
		*client,
	}, nil
}

func (client *Bl3VipClient) Login(username string, password string)	error {
	data := map[string]string {
		"username": username,
		"password": password,
	}

	loginRes, err := client.PostJson("https://api.2k.com/borderlands/users/authenticate", data)
	if err != nil {
		return errors.New("Failed to submit login credentials")
	}
	defer loginRes.Body.Close()

	if loginRes.StatusCode != 200 {
		return errors.New("Failed to login")
	}

	if loginRes.Header.Get("X-CT-REDIRECT") == "" {
		return errors.New("Failed to start session")
	}

	sessionRes, err := client.Get(loginRes.Header.Get("X-CT-REDIRECT")) 
	if err != nil {
		return errors.New("Failed to get session")
	}
	defer sessionRes.Body.Close()

	return nil
}

func (client *Bl3VipClient) GetFullCodeMap() (VipCodeMap, error) {
	codeMap := NewVipCodeMap()
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
	
	redditHtml.Find("[data-test-id='post-content'] tbody tr").Each(func (i int, row *goquery.Selection) {
		if len(row.Find("td").Nodes) < 4 {
			return
		}

		codeTypes := ""
		code := ""

		row.Find("td").EachWithBreak(func (i int, col *goquery.Selection) bool {
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
		
		for _, codeType := range GetCodeTypesInString(codeTypes) {
			codeMap[codeType].Add(code)
		}
	})
	return codeMap, nil
}

func (client *Bl3VipClient) GetRedeemedCodeMap() (VipCodeMap, error) {
	codeMap := NewVipCodeMap()

	url := "https://2kgames.crowdtwist.com/request?widgetId=9470"
	data := map[string]interface{}{
		"model_data": map[string]interface{}{
			"activity": map[string]interface{}{
				"newest_activities": map[string]interface{}{
					"properties": []string{"notes", "title"},
					"query": map[string]interface{}{ 
						"type": "user_activities_me", 
						"args":  map[string]int{
							"row_start": 1,
							"row_end": 1000000,
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
		CodeTypes  string     `json:"title"`
		Code  string          `json:"notes"`
	}

	activities := make([]activity, 0)
	resJson, err := res.BodyAsJson()
	if err != nil {
		return codeMap, err
	}

	resJson.From("model_data.activity.newest_activities").Out(&activities)
	
	for _, act := range activities {
		for _, codeType := range GetCodeTypesInString(act.CodeTypes) {
			codeMap.Add(codeType, act.Code)
		}
	}
    return codeMap, nil
}

func (client *Bl3VipClient) getWidgetConf(url string) *gojsonq.JSONQ {
	response, err := client.Get(url)
	if err != nil {
		return nil
	}

	widgetHtml, err := response.BodyAsHtmlDoc()
	if err != nil {
		return nil
	}

	script := ""

	widgetHtml.Find("script").EachWithBreak(func (i int, scriptTag *goquery.Selection) bool {
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

func (client *Bl3VipClient) GetCodeTypeUrlMap() (VipCodeTypeMap, error) {
	codeTypeUrlMap := NewVipCodeTypeMap()

	widgetConf := client.getWidgetConf("https://2kgames.crowdtwist.com/widgets/t/activity-list/9904/?__locale__=en#2")
	if widgetConf == nil {
		return codeTypeUrlMap, errors.New("Failed to get code redemption types")
	}

	type widget struct {
		WidgetId  int        `json:"widgetId"`
		WidgetName  string   `json:"widgetName"`
	}

	widgets := make([]widget, 0)
	widgetConf.From("entries").Select("link.widgetId", "link.widgetName").Out(&widgets)
	
	for _, wid := range widgets {
		for _, codeType := range GetCodeTypesInString(wid.WidgetName) {
			codeTypeUrlMap[codeType] = "https://2kgames.crowdtwist.com/widgets/t/code-redemption/" + strconv.Itoa(wid.WidgetId)
		}
	}

	return codeTypeUrlMap, nil
}

func (client *Bl3VipClient) GetCodeRedemptionUrl(url string) (string, error) {
	widgetConf := client.getWidgetConf(url)
	if widgetConf == nil {
		return "", errors.New("Failed to get code redemption url")
	}

	campaignId, ok := widgetConf.Find("campaignId").(float64)
	if !ok {
		return "", errors.New("Failed to get code redemption id")
	}
	return "https://2kgames.crowdtwist.com/code-redemption-campaign/redeem?cid=" + strconv.Itoa(int(campaignId)), nil
}