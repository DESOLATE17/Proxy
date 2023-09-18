package usecase

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy/internal/models"
	"proxy/internal/pkg/api"
	"strings"
)

type Usecase struct {
	repo api.Repo
}

func NewUsecase(repo api.Repo) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) SaveRequest(request *http.Request) (int, error) {
	var requestParsed = models.Request{
		Method:    request.Method,
		Scheme:    request.URL.Scheme,
		Host:      request.Host,
		Path:      request.URL.Path,
		Cookies:   u.getCookies(request.Cookies()),
		Headers:   u.getHeaders(request.Header),
		GetParams: u.getRequestParams(request),
	}

	if request.Method == "POST" && request.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		requestParsed.PostParams = u.getPostParams(request)
	}
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return 0, err
	}

	requestParsed.Body = string(bodyBytes)

	return u.repo.SaveRequest(requestParsed)
}

func (u *Usecase) SaveResponse(requestId int, response *http.Response) (models.Response, error) {
	var responseParsed = models.Response{
		RequestId: requestId,
		Code:      response.StatusCode,
		Cookies:   u.getCookies(response.Cookies()),
		Headers:   u.getHeaders(response.Header),
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return models.Response{}, err
	}

	responseParsed.Body = string(bodyBytes)
	err = u.repo.SaveResponse(requestId, responseParsed)
	return responseParsed, err
}

func (u *Usecase) getCookies(cookies []*http.Cookie) []models.Cookies {
	arrCookies := make([]models.Cookies, 0, len(cookies))
	for _, v := range cookies {
		arrCookies = append(arrCookies, models.Cookies{
			Key:   v.Name,
			Value: v.Value,
		})
	}

	return arrCookies
}

func (u *Usecase) getPostParams(request *http.Request) []models.Param {
	err := request.ParseForm()
	if err != nil {
		log.Println("Failed to parse form:", err)
		return []models.Param{}
	}
	arrParams := make([]models.Param, 0, len(request.PostForm))

	for paramName, values := range request.PostForm {
		for _, value := range values {
			arrParams = append(arrParams, models.Param{
				Key:   paramName,
				Value: value,
			})
		}
	}

	return arrParams

}

func (u *Usecase) getResponseCookies(request *http.Request) []models.Cookies {
	arrCookies := make([]models.Cookies, 0, len(request.Cookies()))
	for _, v := range request.Cookies() {
		arrCookies = append(arrCookies, models.Cookies{
			Key:   v.Name,
			Value: v.Value,
		})
	}

	return arrCookies
}

func (u *Usecase) getHeaders(headers map[string][]string) string {
	var stringHeaders string
	for key, values := range headers {
		for _, value := range values {
			stringHeaders += key + " " + value + "\n"
		}
	}
	return stringHeaders
}

func (u *Usecase) getRequestParams(request *http.Request) []models.Param {
	arrParams := make([]models.Param, 0, len(request.URL.Query()))
	for paramName, values := range request.URL.Query() {
		for _, value := range values {
			arrParams = append(arrParams, models.Param{
				Key:   paramName,
				Value: value,
			})
		}
	}

	return arrParams
}

func (u *Usecase) convertStringToHeaders(headersString string) map[string]string {
	headers := make(map[string]string)

	lines := strings.Split(headersString, "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, " ")
			key := parts[0]
			value := parts[1]
			headers[key] = value
		}
	}

	return headers
}

func (u *Usecase) AllRequests() ([]models.Request, error) {
	return u.repo.AllRequests()
}

func (u *Usecase) GetRequest(id int) (models.Request, error) {
	return u.repo.GetRequest(id)
}

func (u *Usecase) RepeatRequest(id int) (models.Response, error) {
	request, err := u.GetRequest(id)
	if err != nil {
		return models.Response{}, err
	}

	body := bytes.NewBufferString(request.Body)
	urlStr := request.Scheme + "://" + request.Host + request.Path
	for i, v := range request.GetParams {
		if i == 0 {
			urlStr += "?"
		}
		urlStr += v.Key + "=" + v.Value
	}

	req, err := http.NewRequest(request.Method, urlStr, body)

	if err != nil {
		fmt.Println(err)
		return models.Response{}, err
	}

	for key, value := range u.convertStringToHeaders(request.Headers) {
		req.Header.Add(key, value)
	}

	reqId, err := u.SaveRequest(req)
	if err != nil {
		log.Printf("Error save: %v", err)
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return models.Response{}, err
	}
	defer resp.Body.Close()

	res, err := u.SaveResponse(reqId, resp)
	if err != nil {
		log.Printf("Error save: %v", err)
	}

	return res, nil
}
