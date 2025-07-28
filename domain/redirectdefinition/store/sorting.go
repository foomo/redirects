package redirectstore

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type SortField string

const (
	SortFieldSource        SortField = "source"
	SortFieldUpdated       SortField = "updated"
	SortFieldLastUpdatedBy SortField = "lastUpdatedBy"
)

type Direction string

const (
	DirectionAscending  Direction = "ascending"
	DirectionDescending Direction = "descending"
)

type Sort struct {
	Field     SortField `json:"field"`
	Direction Direction `json:"direction"`
}

func (d Direction) GetSortValue() int {
	switch d {
	case DirectionAscending:
		return 1
	case DirectionDescending:
		return -1
	default:
		return 1
	}
}
