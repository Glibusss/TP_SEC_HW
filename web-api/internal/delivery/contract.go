package delivery

import (
	"proxy-server/pkg/domain"
)

type Repository interface {
	GetAll() ([]domain.HTTPTransaction, error)
	GetByID(string) (domain.HTTPTransaction, error)
}
