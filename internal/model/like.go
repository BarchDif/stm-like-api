package streaming

import (
	"github.com/google/uuid"
	"time"
)

type Like struct {
	Id        uint64    `json:"id"`
	StreamId  uint64    `json:"stream_id"`
	Uuid      uuid.UUID `json:"uuid"`
	Timestamp time.Time `json:"timestamp"`
}

type EventType uint8

type EventStatus uint8

const (
	Created EventType = iota
	Updated
	Removed

	Deferred EventStatus = iota
	Processed
)

type LikeEvent struct {
	ID     uint64
	Type   EventType
	Status EventStatus
	Entity *Like
}
