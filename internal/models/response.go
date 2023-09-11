package models

type Response struct {
	Id        int               `json:"id"`
	RequestId int               `json:"request_id"`
	Code      int               `json:"code"`
	Message   string            `json:"message"`
	Cookies   string            `json:"cookies"`
	Header    map[string]string `json:"header"`
	Body      string            `json:"body"`
}
