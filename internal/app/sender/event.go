package sender

import (
	"github.com/BarchDif/stm-like-api/internal/model"
)

type EventSender interface {
	Send(subdomain *model.SubdomainEvent) error
}
