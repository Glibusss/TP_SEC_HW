package delivery

import (
	"proxy-server/pkg/domain"
)

type Repository interface {
	Add(domain.HTTPTransaction) error
}
