//go:generate mockgen -destination=../../mocks/sender_mock.go -package=mocks github.com/BarchDif/stm-like-api/internal/app/sender EventSender

package sender

import (
	"github.com/BarchDif/stm-like-api/internal/model"
)

type EventSender interface {
	Send(subdomain *streaming.LikeEvent) error
}
