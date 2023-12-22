package jobs

import (
	"reflect"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"tianxian.com/tee-signer-core/common"
	"tianxian.com/tee-signer-core/log"
)

const (
	refreshInterval = 1
	expiredInterval = 3
)

const minDuration time.Duration = -1 << 63

type Worker struct {
	db *gorm.DB

	Chain         string
	EmptyInterval int64
	ErrInterval   int64

	lastBlockID uint64
	lastTxID    uint64
	lastEventID uint64
}

func NewWorker(conf Bean, gdb *gorm.DB) *Worker {
	return &Worker{
		db:            gdb,
		Chain:         conf.Chain,
		EmptyInterval: conf.EmptyInterval,
		ErrInterval:   conf.ErrInterval,
	}
}

func (w *Worker) NewLoop(
	loopFunc func() (bool, error),
) func(chan struct{}) {
	funcName := runtime.FuncForPC(reflect.ValueOf(loopFunc).Pointer()).Name()
	return func(shutdown chan struct{}) {
		timer := time.NewTimer(minDuration)
		for {
			select {
			case <-shutdown:
				log.Entry.WithField("chain", w.Chain).Warnf("%s stop work", funcName)
				return
			case <-timer.C:
				empty, err := loopFunc()
				if err != nil {
					log.Entry.WithField("chain", w.Chain).WithError(err).Error(err)
					timer.Reset(time.Millisecond * time.Duration(w.ErrInterval))
				} else if empty {
					timer.Reset(time.Millisecond * time.Duration(w.EmptyInterval))
				} else {
					timer.Reset(minDuration)
				}
			}
		}
	}
}

func GetJobs() []func(chan struct{}) {
	result := make([]func(chan struct{}), 0, 3*len(Conf.Bean))
	for _, item := range Conf.Bean {
		singletonHandler, err := common.NewSingletonHandler(
			"pump",
			item.DBConnStr,
			refreshInterval, expiredInterval)
		if err != nil {
			log.Entry.WithError(err).Fatal(err)
		}

		db, err := common.OpenGorm(item.DBConnStr, logrus.Fields{"chain": item.Chain})
		if err != nil {
			log.Entry.WithError(err).Fatal(err)
		}

		w := NewWorker(item, db)

		result = append(result,
			singletonHandler.Loop,
			w.NewLoop(w.loopBlock),
			w.NewLoop(w.loopTransaction),
			w.NewLoop(w.loopEvent),
		)
	}
	initKafkaProducer()
	return result
}
