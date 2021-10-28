//go:generate mockgen -destination=../../mocks/repo_mock.go -package=mocks github.com/BarchDif/stm-like-api/internal/app/repo EventRepo

package repo

import (
	"github.com/BarchDif/stm-like-api/internal/model"
)

type EventRepo interface {
	// Пример запроса на sqlite
	//	select min(Id) as EventId
	//	from LikeEvent
	//	group by LikeId
	//	having max(case Status when 'locked' then 1 else 0 end) = 0
	//	order by EventId
	//	limit 5
	Lock(n uint64) ([]streaming.LikeEvent, error)
	Unlock(eventIDs []uint64) error

	Add(event []streaming.LikeEvent) error
	Remove(eventIDs []uint64) error
}
