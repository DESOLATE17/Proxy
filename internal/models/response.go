package models

type Response struct {
	Id            int       `json:"id"`
	ContentLength int64     `json:"content_length"`
	RequestId     int       `json:"request_id"`
	Code          int       `json:"code"`
	Message       string    `json:"message"`
	Cookies       []Cookies `json:"cookies"`
	Headers       string    `json:"headers"`
	Body          string    `json:"body"`
}

const ConnectionEstablished = "HTTP/1.1 200 Connection established\r\n\r\n"
