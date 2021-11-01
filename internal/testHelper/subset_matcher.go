package testHelper

import (
	"fmt"
	streaming "github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
)

type subsetMatcher struct {
	data []streaming.LikeEvent
}

func IsSubset(data []streaming.LikeEvent) gomock.Matcher {
	return &subsetMatcher{
		data: data,
	}
}

func (m subsetMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case []uint64:
		return m.matchByEventId(x.([]uint64))
	case *streaming.LikeEvent:
		return m.matchByEvent(x.(*streaming.LikeEvent))
	default:
		return false
	}
}

func (m subsetMatcher) String() string {
	return fmt.Sprint("Should be subset of ", m.data)
}

func (m subsetMatcher) matchByEventId(idList []uint64) bool {
	idTable := make(map[uint64]struct{})
	for _, event := range m.data {
		idTable[event.ID] = struct{}{}
	}

	for _, id := range idList {
		if _, ok := idTable[id]; !ok {
			return false
		}
	}

	return true
}

func (m subsetMatcher) matchByEvent(event *streaming.LikeEvent) bool {
	for _, testEvent := range m.data {
		if testEvent == *event {
			return true
		}
	}

	return false
}
