package storage

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	sdkconfig "gitlab.newhuoapps.com/dcenter/mpc-sdk/config"
	"gitlab.newhuoapps.com/dcenter/mpc-sdk/features"

	// "github.com/edgelesssys/ego/ecrypto"

	"gitlab.newhuoapps.com/dcenter/mpc-custody/common/crypto/aes"
)

type JsonDB struct {
	sync.RWMutex
	FilePath      string
	DBData        DBDataType
	BKey          []byte
	EncryptionKey []byte
}

func OpenBadgerDB(cfg sdkconfig.Storage) (features.Storage, error) {
	var err error
	var bKey, encryptionKey []byte

	if _, err = os.Stat(cfg.DBFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("openbadgerdb json database not existed")
	}
	storage := &JsonDB{
		FilePath:      cfg.DBFilePath,
		BKey:          bKey,
		EncryptionKey: encryptionKey,
	}
	data, err := storage.read()
	var dbData DBDataType
	err = json.Unmarshal(data, &dbData)
	storage.DBData = dbData
	return storage, nil
}

func CreateBadgerDB(cfg sdkconfig.Storage) (features.Storage, error) {
	if _, err := os.Stat(cfg.DBFilePath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("createbadgerdb json database already existed")
	}
	var bKey, encryptionKey []byte
	return &JsonDB{
		FilePath:      cfg.DBFilePath,
		DBData:        DBDataType{},
		BKey:          bKey,
		EncryptionKey: encryptionKey,
	}, nil
}

func (b *JsonDB) read() ([]byte, error) {
	data, err := os.ReadFile(b.FilePath)
	if err != nil {
		return nil, fmt.Errorf("read file error: %s", err.Error())
	}
	// sgxData, err := ecrypto.Unseal(data, []byte("sinohope"))
	// if err != nil {
	// 	return nil, fmt.Errorf("sgx unseal file failed, %v", err)
	// }
	return data, nil
}

func (b *JsonDB) write() error {
	b.Lock()
	defer b.Unlock()
	data, err := json.Marshal(b.DBData)
	if err != nil {
		return fmt.Errorf("write jsondb file failed, %v", err)
	}
	// sgxData, err := ecrypto.SealWithProductKey(data, []byte("sinohope"))
	// if err != nil {
	// 	return fmt.Errorf("sgx seal file failed, %v", err)
	// }
	return os.WriteFile(b.FilePath, data, 0644)
}

func (b *JsonDB) Flush() error {
	return b.write()
}

func (b *JsonDB) Set(key string, value []byte) error {
	err := b.DBData.Set(key, value)
	if err != nil {
		return fmt.Errorf("update dbdata failed, %v", err)
	}
	err = b.write()
	return err
}

func (b *JsonDB) Get(key string) ([]byte, error) {
	return b.DBData.Get(key)
}

func (b *JsonDB) Delete(key string) error {
	err := b.DBData.Delete(key)
	if err != nil {
		return err
	}
	err = b.write()
	return err
}

func (b *JsonDB) Start() error {
	return nil
}

func (b *JsonDB) Stop() error {
	b.BKey = nil
	b.EncryptionKey = nil
	return b.write()
}

func (b *JsonDB) TryUnseal(password []byte) error {
	if len(b.BKey) > 0 && len(b.EncryptionKey) > 0 {
		// Already unsealed
		return nil
	}
	bKey, err := password2BKey(password)
	if err != nil {
		return fmt.Errorf("tryunseal json generate bKey failed, %v", err)
	}
	encryptionKey, err := loadEncryptionKey(b, bKey)
	if err != nil {
		return fmt.Errorf("tryunseal json load encryption key failed, %v", err)
	}
	b.BKey = bKey
	b.EncryptionKey = encryptionKey
	return nil
}

func (b *JsonDB) IsUnsealed() bool {
	b.RLock()
	defer b.RUnlock()
	if len(b.BKey) > 0 && len(b.EncryptionKey) > 0 {
		// Already unsealed
		return true
	}
	return false
}

func (b *JsonDB) Encrypt(data string) (string, error) {
	cipherData, err := aes.Encrypt([]byte(data), b.EncryptionKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(cipherData), nil
}

func (b *JsonDB) Decrypt(data string) (string, error) {
	cipherData, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	plainData, err := aes.Decrypt(cipherData, b.EncryptionKey)
	if err != nil {
		return "", err
	}
	return string(plainData), nil
}

func (b *JsonDB) GetCipherEK() (string, error) {
	cipherEK, err := aes.Encrypt(b.BKey, b.EncryptionKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(cipherEK), nil
}
