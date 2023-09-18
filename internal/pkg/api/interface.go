package api

import (
	"net/http"
	"proxy/internal/models"
)

type Repo interface {
	SaveRequest(request models.Request) (int, error)
	AllRequests() ([]models.Request, error)
	SaveResponse(requestId int, response models.Response) error
	GetRequest(id int) (models.Request, error)
}

type Usecase interface {
	SaveRequest(request *http.Request) (int, error)
	SaveResponse(requestId int, response *http.Response) (models.Response, error)
	AllRequests() ([]models.Request, error)
	GetRequest(id int) (models.Request, error)
	RepeatRequest(id int) (models.Response, error)
}
