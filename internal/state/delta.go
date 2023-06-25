package state

type ZoneDelta struct {
	Insert ZoneCollection
	Update ZoneCollection
	Delete ZoneCollection
}

func NewZoneDelta() *ZoneDelta {
	return &ZoneDelta{
		Insert: ZoneCollection{},
		Update: ZoneCollection{},
		Delete: ZoneCollection{},
	}
}

func (z *ZoneDelta) TotalOperations() int {
	return len(z.Insert) + len(z.Update) + len(z.Delete)
}

func (z *ZoneDelta) TotalInsertUpdates() int {
	return len(z.Insert) + len(z.Update)
}

func (z *ZoneDelta) InsertUpdate() ZoneCollection {
	zc := ZoneCollection{}
	zc = append(zc, z.Insert...)
	zc = append(zc, z.Update...)
	return zc
}
