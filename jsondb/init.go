package jsondb

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"

	"gitlab.newhuoapps.com/dcenter/mpc-custody/common/crypto/aes"
)

func password2BKey(bKey []byte) ([]byte, error) {
	if len(bKey) == 0 {
		return nil, fmt.Errorf("password is empty")
	}
	return argon2.IDKey(bKey, nil, 1, 128*1024, 2, 32), nil
}

func loadEncryptionKey(db *JsonDB, bKey []byte) ([]byte, error) {
	encryptionKey, err := read(db, bKey, dbPathEncryptionKey)
	if err != nil {
		logrus.Debugf("loadEncryptionKey json error %v", err)
		if !strings.Contains(err.Error(), EncryptionKeyNotFound) {
			return nil, err
		}
		encryptionKey, err = newEncryptionKey()
		if err != nil {
			return nil, err
		}
		err = write(db, bKey, dbPathEncryptionKey, encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	return encryptionKey, nil
}

func read(db *JsonDB, bKey []byte, path string) ([]byte, error) {
	value, err := db.Get(path)
	if err != nil {
		logrus.Debugf("read json error %v", err)
		return nil, err
	}
	cipherData := []byte(value)
	// logrus.Debugf("read aes  %s %s", value, string(bKey))
	plainData, err := aes.Decrypt(cipherData, bKey)
	if err != nil {
		logrus.Debugf("read json aes error %v", err)
		return nil, err
	}
	return plainData, err
}

func write(db *JsonDB, bKey []byte, path string, data []byte) error {
	cipherData, err := aes.Encrypt(data, bKey)
	if err != nil {
		return err
	}
	err = db.Set(path, cipherData)
	if err != nil {
		return err
	}
	return nil
}

func newEncryptionKey() ([]byte, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
