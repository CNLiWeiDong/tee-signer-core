package jsondb

func (b *JsonDB) Kvv1Read(path string) (string, error) {
	value, err := b.Get(path)
	if err != nil {
		return "", err
	}
	return b.Decrypt(string(value))
}

func (b *JsonDB) Kvv1Write(path string, data string) error {
	cipherData, err := b.Encrypt(data)
	if err != nil {
		return err
	}
	return b.Set(path, []byte(cipherData))
}

func (b *JsonDB) Kvv1Delete(path string) error {
	err := b.Delete(path)
	if err != nil {
		return err
	}
	return nil
}
