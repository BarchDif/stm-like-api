package retranslator

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockEventRepo(ctrl)
	sender := mocks.NewMockEventSender(ctrl)

	rootContext := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(rootContext)

	repo.EXPECT().Lock(gomock.Any()).AnyTimes()

	cfg := Config{
		ChannelSize:    512,
		ConsumerCount:  2,
		ConsumeSize:    10,
		ConsumeTimeout: 10 * time.Second,
		ProducerCount:  2,
		WorkerCount:    2,
		Repo:           repo,
		Sender:         sender,
	}

	retranslator := NewRetranslator(cfg)

	t.Run("StartsSuccessfully", func(t *testing.T) {
		retranslator.Start(cancelCtx)
	})

	t.Run("StopsSuccessfully", func(t *testing.T) {
		cancelFunc()
		<-retranslator.Stopped()
	})
}
