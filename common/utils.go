package common

import (
	"runtime"
	"time"

	"blockchain.com/pump/log"
)

// TimeConsume provides convenience function for time-consuming calculation
func TimeConsume(start time.Time) {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return
	}

	// get Fun object from pc
	funcName := runtime.FuncForPC(pc).Name()
	log.Entry.WithField("tags", "func_time_consume").
		WithField("cost", time.Since(start).String()).Debug(funcName)
}

// CheckList 检查是否包含元素
type CheckList []string

func (l *CheckList) Contains(value string) bool {
	for _, item := range *l {
		if item == value {
			return true
		}
	}
	return false
}
