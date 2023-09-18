package models

type Request struct {
	Id         int       `json:"id"`
	Method     string    `json:"method"`
	Scheme     string    `json:"scheme"`
	Host       string    `json:"host"`
	Path       string    `json:"path"`
	Cookies    []Cookies `json:"cookies"`
	Headers    string    `json:"headers,omitempty"`
	Body       string    `json:"body,omitempty"`
	GetParams  []Param   `json:"get_params"`
	PostParams []Param   `json:"post_params"`
}

type Cookies struct {
	Key   string `json:"cookie_name"`
	Value string `json:"cookie_value"`
}

type Param struct {
	Key   string `json:"param_name"`
	Value string `json:"param_value"`
}
