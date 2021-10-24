package repo

import (
	"github.com/BarchDif/stm-like-api/internal/model"
)

type EventRepo interface {
	Lock(n uint64) ([]model.SubdomainEvent, error)
	Unlock(eventIDs []uint64) error

	Add(event []model.SubdomainEvent) error
	Remove(eventIDs []uint64) error
}
