package model

type System struct {
	ID        int64  `db:"id"`
	System    string `db:"system"`
	Number    int32  `db:"number"`
	Active    bool   `db:"active"`
	Size      int32  `db:"size"`
	FishCount int32  `db:"fish_count"`
}
