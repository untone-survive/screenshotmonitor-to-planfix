package sm

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const ApiUrl = "https://screenshotmonitor.com/api/v2/"
const GetActivities = "GetActivities"
const GetScreenshots = "GetScreenshots"

type SM struct {
	Token string
}

type GetActivitiesRequest []GetActivitiesRequestItem

type GetActivitiesRequestItem struct {
	Employee int   `json:"employmentId"`
	From     int64 `json:"from"`
	To       int64 `json:"to"`
}

type GetActivityResponse []*GetActivityResponseItem

type GetActivityResponseItem struct {
	ActivityId    string `json:"activityId"`
	PlanfixId     int    `json:"-"`
	PlanfixRealId int    `json:"-"`
	ScreenshotUrl string `json:"-"`
	Id            string `json:"id"`
	EmploymentId  int    `json:"employmentId"`
	Note          string `json:"note"`
	Offline       bool   `json:"offline"`
	From          int64  `json:"from"`
	To            int64  `json:"to"`
	ProjectId     string `json:"projectId"`
}

func New(token string) *SM {
	return &SM{Token: token}
}

func (s *SM) GetActivities(req GetActivitiesRequest) (*GetActivityResponse, error) {
	resp := new(GetActivityResponse)
	err := s.request(GetActivities, req, resp)
	return resp, err
}

type GetScreenshotsRequest []uuid.UUID

type GetScreenshotsResponse []GetScreenshotResponseItem

type GetScreenshotResponseItem struct {
	Id            int    `json:"id"`
	ActivityId    string `json:"activityId"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Url           string `json:"url"`
	ThumbUrl      string `json:"thumbUrl"`
	Taken         int64  `json:"taken"`
	ActivityLevel int    `json:"activityLevel"`
	Applications  []struct {
		FromScreen      bool   `json:"fromScreen"`
		Duration        int    `json:"duration"`
		ApplicationName string `json:"applicationName"`
	} `json:"applications"`
}

func (s *SM) GetScreenshots(req GetScreenshotsRequest) (*GetScreenshotsResponse, error) {
	resp := new(GetScreenshotsResponse)
	err := s.request(GetScreenshots, req, resp)
	return resp, err
}

func (s *SM) request(method string, requestStruct, responseStruct interface{}) error {
	jsonBytes, err := json.Marshal(requestStruct)
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	httpClient := http.Client{}
	req, _ := http.NewRequest("POST", ApiUrl+"/"+method, strings.NewReader(string(jsonBytes)))
	req.Header.Add("X-SSM-Token", s.Token)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Network error while request to planfix: %s", err)
		return err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)

	json.Unmarshal(respData, &responseStruct)

	return nil
}

func (s *SM) NewRequestActivityForUser(userId int, startDate, endDate time.Time) GetActivitiesRequest {
	return GetActivitiesRequest{
		{
			Employee: userId,
			From:     startDate.Unix(),
			To:       endDate.Unix(),
		},
	}
}

func (s *SM) NewRequestScreenshotsArgs(activityId string) GetScreenshotsRequest {
	activityUuid, err := uuid.Parse(activityId)
	if err != nil {
		activityUuid = uuid.Nil
	}
	return GetScreenshotsRequest{
		activityUuid,
	}
}

func (a GetActivityResponseItem) GetFrom() time.Time {
	return time.Unix(a.From, 0)
}

func (a GetActivityResponseItem) GetTo() time.Time {
	return time.Unix(a.To, 0)
}

//FromFormatted retrieve formatted `From` field
func (a GetActivityResponseItem) FromFormatted() string {
	return a.GetFrom().Format("02.01.2006")
}

//FromFormattedTime retrieve formatted time of `From` field
func (a GetActivityResponseItem) FromFormattedTime() string {
	return a.GetFrom().Format("15:04")
}

//ToFormatted retrieve formatted `To` field
func (a GetActivityResponseItem) ToFormatted() string {
	return a.GetTo().Format("02.01.2006")
}

//ToFormattedTime retrieve formatted time of `Time` field
func (a GetActivityResponseItem) ToFormattedTime() string {
	return a.GetTo().Format("15:04")
}
