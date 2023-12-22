package jobs

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"tianxian.com/tee-signer-core/common"
	"tianxian.com/tee-signer-core/model"
)

func (w *Worker) loopTransaction() (bool, error) {
	defer common.TimeConsume(time.Now())
	id, err := w.getLastTxID()
	if err != nil {
		return false, err
	}

	batchSize := 10
	list := make([]*Transaction, 0, batchSize)
	err = w.db.
		Where("id > ?", id).
		Limit(batchSize).
		Order("id").
		Find(&list).
		Error
	if err != nil {
		return false, errors.WithStack(err)
	}
	if len(list) == 0 {
		return true, nil
	}

	for _, tx := range list {
		if tx.Result == TransactionResultSuccess {
			// 此处的赋值, 可以避免程序在后续步骤报错时, 重复执行
			w.lastTxID = tx.ID
			continue
		}
		b := new(Block)
		err = w.db.Where("id = ?", tx.BlockID).Take(b).Error
		if err != nil {
			return false, errors.WithStack(err)
		}
		err = sendTx2Kafka(b, tx)
		if err != nil {
			return false, err
		}
	}

	id = list[len(list)-1].ID
	err = w.db.
		Model(&InitConfig{}).
		Where("name = ?", model.AnchorTx).
		Update("value", id).
		Error
	if err != nil {
		return false, err
	}
	w.lastTxID = id
	return len(list) < batchSize, nil
}

// 获取锚点
func (w *Worker) getLastTxID() (uint64, error) {
	if w.lastTxID != 0 {
		return w.lastTxID, nil
	}
	item := new(InitConfig)
	err := w.db.Where("name = ?", model.AnchorTx).Take(item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.Name = model.AnchorTx
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
	w.lastTxID = value
	return value, nil
}

func sendTx2Kafka(b *Block, tx *Transaction) error {
	topic := Conf.Business[tx.BizType]
	data := NewKafkaTransaction(b, tx)
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}
	return send2Kafka(topic, bs)
}
