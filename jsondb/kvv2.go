package jsondb

func (b *JsonDB) Kvv2Read(path string) (string, error) {
	return b.Kvv1Read(path)
}

func (b *JsonDB) Kvv2Write(path string, data string) error {
	return b.Kvv1Write(path, data)
}

func (b *JsonDB) Kvv2Delete(path string) error {
	return b.Kvv1Delete(path)
}
