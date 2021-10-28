package testdata

import streaming "github.com/BarchDif/stm-like-api/internal/model"

const eventCount = 20

var Events []streaming.LikeEvent

func init() {
	Events = make([]streaming.LikeEvent, 0, eventCount)

	for i := 0; i < eventCount; i++ {
		Events = append(Events, streaming.LikeEvent{ID: uint64(i)})
	}
}
