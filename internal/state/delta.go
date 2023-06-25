package state

type ZoneDelta struct {
	Insert []Zone
	Update []Zone
	Delete []Zone
}

func NewZoneDelta() *ZoneDelta {
	return &ZoneDelta{
		Insert: []Zone{},
		Update: []Zone{},
		Delete: []Zone{},
	}
}

func (z *ZoneDelta) TotalOperations() int {
	return len(z.Insert) + len(z.Update) + len(z.Delete)
}

func (z *ZoneDelta) TotalInsertUpdates() int {
	return len(z.Insert) + len(z.Update)
}

func (z *ZoneDelta) InsertUpdate() []Zone {
	zc := []Zone{}
	zc = append(zc, z.Insert...)
	zc = append(zc, z.Update...)
	return zc
}
