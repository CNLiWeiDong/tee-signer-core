package jsondb

import "fmt"

const (
	EncryptionKeyNotFound = "not found"
)

const (
	dbPathKeyPrefix        = "data"
	dbPathEncryptionKey    = "encryption-key"
	dbPathEncryptionKeyPre = "encryption-key-pre"
)

type DBDataType struct {
	Data    []DBKeyValue `json:"data"`
	Version string       `json:"version"`
}

type DBKeyValue struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

func (db *DBDataType) Get(key string) ([]byte, error) {
	absKey := fmt.Sprintf("%s/%s", dbPathKeyPrefix, key)
	for _, v := range db.Data {
		if v.Key == absKey {
			return v.Value, nil
		}
	}
	return nil, fmt.Errorf("value not found with the key: %s", key)
}

func (db *DBDataType) Set(key string, value []byte) error {
	absKey := fmt.Sprintf("%s/%s", dbPathKeyPrefix, key)
	for i, v := range db.Data {
		if v.Key == absKey {
			db.Data[i].Value = value
			return nil
		}
	}
	db.Data = append(db.Data, DBKeyValue{
		Key:   absKey,
		Value: value,
	})
	return nil
}

func (db *DBDataType) Delete(key string) error {
	absKey := fmt.Sprintf("%s/%s", dbPathKeyPrefix, key)
	for i, v := range db.Data {
		if v.Key == absKey {
			db.Data = append(db.Data[:i], db.Data[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("delete value not found with the key: %s", key)
}
