package links

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Links struct {
	Code   string
	ApiUrl string
}

type ShortenRequest struct {
	Url       Url
	Action    string
	Signature string
	Format    string
	Keyword   string
	Title     string
}

type ShortenResponse struct {
	URL        URLInfo `json:"url"`
	Status     string  `json:"status"`
	Message    string  `json:"message"`
	Title      string  `json:"title"`
	ShortURL   string  `json:"shorturl"`
	StatusCode string  `json:"statusCode"`
}

type URLInfo struct {
	Keyword string `json:"keyword"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	IP      string `json:"ip"`
}

type Url string

func New(code string, api_url string) *Links {
	return &Links{Code: code, ApiUrl: api_url}
}

func (l *Links) Shorten(fullurl Url, title string) (ShortenResponse, error) {
	var linksResp ShortenResponse

	r := ShortenRequest{Url: fullurl, Action: "shorturl", Signature: l.Code, Format: "json"}

	// Build query string
	queryParams := url.Values{}
	queryParams.Set("url", string(r.Url))
	queryParams.Set("action", r.Action)
	queryParams.Set("signature", r.Signature)
	queryParams.Set("format", r.Format)
	if title != "" {
		queryParams.Set("title", title)
	}

	// Create full URL
	apiUrl := fmt.Sprintf("%s?%s", l.ApiUrl, queryParams.Encode())

	// Create HTTP request
	httpClient := http.Client{}
	req, _ := http.NewRequest("POST", apiUrl, nil)
	req.Header.Add("User-Agent", "elustro-sm-planfix-bot/1.0")
	req.Header.Set("Content-Length", "0") // Explicitly set Content-Length

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Network error while request to shortener: %s", err)
		return linksResp, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	stringData := string(data)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return linksResp, fmt.Errorf("[ERROR] status error: %v\n%q", resp.StatusCode, stringData)
	}

	err = json.Unmarshal(data, &linksResp)
	return linksResp, err
}
