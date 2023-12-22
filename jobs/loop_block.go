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

func (w *Worker) loopBlock() (bool, error) {
	defer common.TimeConsume(time.Now())
	id, err := w.getLastBlockID()
	if err != nil {
		return false, err
	}

	item := new(Block)

	err = w.db.Where("state = ?", model.BlockStateConfirmed).Last(item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, errors.WithStack(err)
	}
	if item.ID <= id {
		return true, nil
	}

	// item.ID > id
	err = sendBlock2Kafka(item)
	if err != nil {
		return false, err
	}
	err = w.db.
		Model(&InitConfig{}).
		Where("name = ?", model.AnchorBlock).
		Update("value", item.ID).
		Error
	if err != nil {
		return false, err
	}
	w.lastBlockID = item.ID
	return false, nil
}

// 获取锚点
func (w *Worker) getLastBlockID() (uint64, error) {
	if w.lastBlockID != 0 {
		return w.lastBlockID, nil
	}
	item := new(InitConfig)
	err := w.db.Where("name = ?", model.AnchorBlock).Take(item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.Name = model.AnchorBlock
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
	w.lastBlockID = value
	return value, nil
}

func sendBlock2Kafka(b *Block) error {
	topic := Conf.Business[BizPayment]
	data := NewKafkaBlock(b)
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}
	return send2Kafka(topic, bs)
}
