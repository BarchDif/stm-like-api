package producer_test

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/producer"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/BarchDif/stm-like-api/internal/tests/fixture"
	"github.com/BarchDif/stm-like-api/internal/tests/testdata"
	"sort"
	"testing"
)

const (
	producerCount = 3
)

func startProducer(t *testing.T) (producer.Producer, chan streaming.LikeEvent, chan streaming.LikeEvent, chan streaming.LikeEvent, chan streaming.LikeEvent) {
	fxt := fixture.NewFixture(t)

	senderResult := make(chan streaming.LikeEvent, len(testdata.Events))
	sender := fxt.EventSenderSendMock(senderResult)

	events := make(chan streaming.LikeEvent, len(testdata.Events))
	for i := 0; i < len(testdata.Events); i++ {
		events <- testdata.Events[i]
	}

	pool := fxt.WorkerPoolSubmitMock()

	unlockResult := make(chan streaming.LikeEvent, len(testdata.Events))
	removeResult := make(chan streaming.LikeEvent, len(testdata.Events))
	repo := fxt.LikeRepoUnlockRemoveMock(unlockResult, removeResult)

	producer := producer.NewKafkaProducer(producerCount, repo, sender, events, pool)
	producer.Start(context.Background())

	return producer, events, senderResult, unlockResult, removeResult
}

func compare(first []streaming.LikeEvent, second []streaming.LikeEvent) bool {
	sort.SliceStable(first, func(i, j int) bool {
		return first[i].ID < first[j].ID
	})
	sort.SliceStable(second, func(i, j int) bool {
		return second[i].ID < second[j].ID
	})

	if len(first) != len(second) {
		return false
	}

	for i := 0; i < len(first); i++ {
		if first[i] != second[i] {
			return false
		}
	}

	return true
}

func TestProducer_EventSending(t *testing.T) {
	producer, events, senderResultChannel, _, _ := startProducer(t)

	close(events)
	<-producer.Cancel()

	senderResult := make([]streaming.LikeEvent, 0)
	close(senderResultChannel)
	for event := range senderResultChannel {
		senderResult = append(senderResult, event)
	}
	if !compare(testdata.Events, senderResult) {
		t.Fail()
	}
}

func TestProducer_EventBatching(t *testing.T) {
	producer, events, _, _, removeResultChan := startProducer(t)

	close(events)
	<-producer.Cancel()

	removeResult := make([]streaming.LikeEvent, 0)
	close(removeResultChan)
	for event := range removeResultChan {
		removeResult = append(removeResult, event)
	}
	if !compare(testdata.Events, removeResult) {
		t.Fail()
	}
}
