package auditorium

type Auditorium struct {
	Id          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Capacity    int    `db:"capacity" json:"capacity"`
	IsAvailable bool   `db:"is_available" json:"isAvailable"`
}
