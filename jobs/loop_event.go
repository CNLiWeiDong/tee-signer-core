package jobs

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"tianxian.com/tee-signer-core/common"
	"tianxian.com/tee-signer-core/log"
	"tianxian.com/tee-signer-core/model"
)

var (
	batchEventList common.CheckList = []string{
		"batchPayPriceV2",
		"setNftWithdrawState",
		"setNftDepositState",
		"withdrawNft",
		"depositNft",
		"setNftType",
	}
)

func (w *Worker) loopEvent() (bool, error) {
	defer common.TimeConsume(time.Now())
	id, err := w.getLastEventID()
	if err != nil {
		return false, err
	}

	list, empty, err := batchNextEvent(w.db, nil, 3, id)
	if empty || err != nil {
		return empty, err
	}

	b := new(Block)
	err = w.db.Where("id = ?", list[0].BlockID).Take(b).Error
	if err != nil {
		return false, errors.WithStack(err)
	}
	tx := new(Transaction)
	err = w.db.Where("id = ?", list[0].TransactionID).Take(tx).Error
	if err != nil {
		return false, errors.WithStack(err)
	}

	checkPoint := time.Now()

	if batchEventList.Contains(tx.FuncName) {
		err = sendEvents2Kafka(w.Chain, b, tx, list)
		if err != nil {
			return false, err
		}
	} else {
		for _, item := range list {
			err = sendEvents2Kafka(w.Chain, b, tx, []*Event{item})
			if err != nil {
				return false, err
			}
		}
	}

	log.Entry.WithField("tags", "event2kafka").Debug(time.Since(checkPoint))

	id = list[len(list)-1].ID
	err = w.db.
		Model(&InitConfig{}).
		Where("name = ?", model.AnchorEvent).
		Update("value", id).
		Error
	if err != nil {
		return false, err
	}
	w.lastEventID = id
	return false, nil
}

// 获取锚点
func (w *Worker) getLastEventID() (uint64, error) {
	if w.lastEventID != 0 {
		return w.lastEventID, nil
	}
	item := new(InitConfig)
	err := w.db.Where("name = ?", model.AnchorEvent).Take(item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.Name = model.AnchorEvent
		item.Value = "0"
		err = w.db.Create(item).Error
		return 0, errors.WithStack(err)
	}
	if err != nil {
		return 0, errors.WithStack(err)
	}

	value, err := strconv.ParseUint(item.Value, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "can not parse %s as uint64", item.Value)
	}
	w.lastEventID = value
	return value, nil
}

func sendEvents2Kafka(chain string, b *Block, tx *Transaction, list []*Event) error {
	topic := Conf.Business[tx.BizType]
	data := NewKafkaEvent(chain, b, tx, list)
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}
	return send2Kafka(topic, bs)
}

// 批量获取下一批 event: transaction_id,state 一致
func batchNextEvent(
	gdb *gorm.DB,
	first *Event,
	limit int,
	lastID uint64) ([]*Event, bool, error) {
	list := make([]*Event, 0, limit)
	err := gdb.
		Where("id > ?", lastID).
		Limit(limit).
		Order("id").
		Find(&list).
		Error
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	if len(list) == 0 {
		return nil, true, nil
	}

	if first == nil {
		first = list[0]
	}

	for idx, item := range list {
		if item.TransactionID != first.TransactionID || item.State != first.State {
			return list[:idx], false, nil
		}
	}

	if len(list) < limit {
		return list, false, nil
	}

	newLastID := list[limit-1].ID

	temp, empty, err := batchNextEvent(gdb, first, limit, newLastID)
	if err != nil {
		return nil, false, err
	}
	if !empty {
		list = append(list, temp...)
	}
	return list, false, nil
}
