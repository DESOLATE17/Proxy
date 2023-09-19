package usecase

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
		RequestId:     requestId,
		ContentLength: response.ContentLength,
		Code:          response.StatusCode,
		Cookies:       u.getCookies(response.Cookies()),
		Headers:       u.getHeaders(response.Header),
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

func (u *Usecase) Scan(id int) (string, error) {
	request, err := u.GetRequest(id)
	if err != nil {
		return "", err
	}

	response, err := u.RepeatRequest(id)
	if err != nil {
		return "", err
	}

	// convert it back to http request
	req, err := u.ConvertModelToRequest(request)
	if err != nil {
		return "", err
	}

	// change GET params
	parsedURL, err := url.Parse(req.URL.String())
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	for key, values := range query {
		for _, value := range values {
			query.Set(key, `'`)
			parsedURL.RawQuery = query.Encode()
			req.URL = parsedURL

			message, err := checkVulnerability(req, response)
			if err != nil || message != "" {
				return message + "GetParam " + key, err
			}

			query.Set(key, `"`)
			parsedURL.RawQuery = query.Encode()
			req.URL = parsedURL

			message, err = checkVulnerability(req, response)
			if err != nil || message != "" {
				return message + "GetParam " + key, err
			}

			query.Set(key, value)
			parsedURL.RawQuery = query.Encode()
			req.URL = parsedURL
		}
	}

	//change POST params
	if req.Method == "POST" {
		for key, values := range req.PostForm {
			for i, val := range values {
				req.PostForm[key][i] = `'`
				message, err := checkVulnerability(req, response)
				if err != nil || message != "" {
					return message + "PostParam " + key, err
				}

				req.PostForm[key][i] = `"`
				message, err = checkVulnerability(req, response)
				if err != nil || message != "" {
					return message + "PostParam " + key, err
				}

				req.PostForm[key][i] = val
			}
		}
	}

	cookies := req.Cookies()
	for i, cookie := range cookies {
		newCookie := &http.Cookie{
			Name:     cookie.Name,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			MaxAge:   cookie.MaxAge,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
			SameSite: cookie.SameSite,
		}

		newCookie.Value = `"`
		cookies[i] = newCookie
		req.Header.Set("Cookie", cookiesToString(cookies))
		message, err := checkVulnerability(req, response)
		if err != nil || message != "" {
			return message + "Cookie " + cookie.Name, err
		}

		cookies[i].Value = `'`
		req.Header.Set("Cookie", cookiesToString(cookies))
		message, err = checkVulnerability(req, response)
		if err != nil || message != "" {
			return message + "Cookie " + cookie.Name, err
		}
	}

	// change HTTP headers
	for key, values := range req.Header {
		for i, value := range values {
			req.Header[key][i] = `'`
			message, err := checkVulnerability(req, response)
			if err != nil || message != "" {
				return message + "Header " + key, err
			}

			req.Header[key][i] = `"`
			message, err = checkVulnerability(req, response)
			if err != nil || message != "" {
				return message + "Header " + key, err
			}

			req.Header[key][i] = value
		}
	}
	return "", nil
}

func cookiesToString(cookies []*http.Cookie) string {
	var str string
	for _, cookie := range cookies {
		str += cookie.String() + "; "
	}
	return strings.TrimRight(str, "; ")
}

func checkVulnerability(req *http.Request, expectedResponse models.Response) (string, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedResponse.Code || resp.ContentLength != expectedResponse.ContentLength {
		return "Уязвимый параметр:", nil
	}
	return "", nil
}

func (u *Usecase) ConvertModelToRequest(request models.Request) (*http.Request, error) {
	httpRequest, err := http.NewRequest(request.Method, request.Scheme+"://"+request.Host+request.Path, strings.NewReader(request.Body))
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range u.convertStringToHeaders(request.Headers) {
		httpRequest.Header.Set(key, value)
	}

	// Set cookies
	for _, cookie := range request.Cookies {
		httpRequest.AddCookie(&http.Cookie{
			Name:  cookie.Key,
			Value: cookie.Value,
		})
	}

	// Set GET parameters
	queryParams := make(url.Values)
	for _, value := range request.GetParams {
		queryParams.Add(value.Key, value.Value)
	}
	httpRequest.URL.RawQuery = queryParams.Encode()

	// Set POST parameters
	if request.Method == "POST" && len(request.PostParams) > 0 {
		formData := make(url.Values)
		for _, param := range request.PostParams {
			formData.Add(param.Key, param.Value)
		}
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpRequest.Body = io.NopCloser(strings.NewReader(formData.Encode()))
	}
	return httpRequest, nil
}
