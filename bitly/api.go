package bitly

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const API_URL = "https://api-ssl.bitly.com/v4/bitlinks"

type Bitly struct {
	Token string
}

type Request struct {
	Url       Url      `json:"long_url"`
	Domain    string   `json:"domain"`
	GroupGuid string   `json:"group_guid,omitempty"`
	Title     string   `json:"title,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Deeplinks []struct {
		AppId       string `json:"app_id,omitempty"`
		AppUriPath  string `json:"app_uri_path,omitempty"`
		InstallUrl  string `json:"install_url,omitempty"`
		InstallType string `json:"install_type,omitempty"`
	} `json:"deeplinks,omitempty"`
}

type Error struct {
	Message     string `json:"message"`
	Resource    string `json:"resource"`
	Description string `json:"description"`
	Errors      []struct {
		Field     string `json:"field"`
		ErrorCode string `json:"error_code"`
	} `json:"errors"`
}

type Response struct {
	References     []interface{} `json:"references"`
	Link           string        `json:"link"`
	Id             string        `json:"id"`
	LongUrl        string        `json:"long_url"`
	Title          string        `json:"title"`
	Archived       bool          `json:"archived"`
	CreatedAt      string        `json:"created_at"`
	CreatedBy      string        `json:"created_by"`
	ClientId       string        `json:"client_id"`
	CustomBitlinks []string      `json:"custom_bitlinks"`
	Tags           []string      `json:"tags"`
	LaunchpadIDs   []string      `json:"launchpad_ids"`
	Deeplinks      []struct {
		Guid        string `json:"guid"`
		Bitlink     string `json:"bitlink"`
		AppUriPath  string `json:"app_uri_path"`
		InstallUrl  string `json:"install_url"`
		AppGuid     string `json:"app_guid"`
		Os          string `json:"os"`
		InstallType string `json:"install_type"`
		Created     string `json:"created"`
		Modified    string `json:"modified"`
		BrandGuid   string `json:"brand_guid"`
	} `json:"deeplinks"`
	IsDeleted bool `json:"is_deleted"`
}

type Url string

func New(token string) *Bitly {
	return &Bitly{Token: token}
}

func (b *Bitly) Shorten(url Url, domain ...string) (Response, error) {
	var bitlyResp Response

	if len(domain) == 0 {
		domain = append(domain, "bit.ly")
	}
	r := Request{Domain: domain[0], Url: url}

	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return bitlyResp, err
	}

	httpClient := http.Client{}
	req, _ := http.NewRequest("POST", API_URL, strings.NewReader(string(jsonBytes)))
	req.Header.Add("User-Agent", "elustro-sm-planfix-bot/1.0")
	req.Header.Add("Authorization", "Bearer "+b.Token)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Network error while request to bit.ly: %s", err)
		return bitlyResp, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	stringData := string(data)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var jsonError Error
		json.Unmarshal(data, &jsonError)
		return bitlyResp, fmt.Errorf("[ERROR] status error: %v\n%q", resp.StatusCode, stringData)
	}

	err = json.Unmarshal(data, &bitlyResp)
	return bitlyResp, err
}
