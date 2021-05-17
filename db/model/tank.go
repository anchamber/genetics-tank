package model

type Tank struct {
	ID        int64  `db:"id"`
	System    string `db:"system"`
	Number    uint32 `db:"number"`
	Active    bool   `db:"active"`
	Size      uint32 `db:"size"`
	FishCount uint32 `db:"fish_count"`
}
