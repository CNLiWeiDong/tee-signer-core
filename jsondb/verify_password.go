package jsondb

import (
	"bytes"
	"fmt"
)

func (b *JsonDB) VerifyPassword(password []byte) error {
	bKey, err := password2BKey(password)
	if err != nil {
		return fmt.Errorf("generate bKey failed, %v", err)
	}
	if !bytes.Equal(bKey, b.BKey) {
		return fmt.Errorf("unknown bkey, all shares are cleaned, should input all of them again")
	}
	return nil
}
