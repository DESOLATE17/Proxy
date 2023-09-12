package usecase

import (
	"proxy/internal/pkg/api"
)

type Usecase struct {
	repo api.Repo
}

func NewUsecase(repo api.Repo) *Usecase {
	return &Usecase{repo: repo}
}
