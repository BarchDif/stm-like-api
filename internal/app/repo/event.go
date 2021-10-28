//go:generate mockgen -destination=../../mocks/repo_mock.go -package=mocks github.com/BarchDif/stm-like-api/internal/app/repo EventRepo

package repo

import (
	"github.com/BarchDif/stm-like-api/internal/model"
)

type EventRepo interface {
	Lock(n uint64) ([]streaming.LikeEvent, error)
	Unlock(eventIDs []uint64) error

	Add(event []streaming.LikeEvent) error
	Remove(eventIDs []uint64) error
}
