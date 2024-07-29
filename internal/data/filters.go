package data

import (
	"math"
	"strings"
)

type Filters struct {
	Page         int `validate:"gte=0,lte=100"`
	PageSize     int `validate:"gte=0,lte=100"`
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage int `json:"current_page,omitempty"`
	PageSize    int `json:"page_size,omitempty"`
	FirstPage   int `json:"first_page,omitempty"`
	LastPage    int `json:"last_page,omitempty"`
	Total       int `json:"total_records,omitempty"`
}

func calculateMetadata(totalrecords int, page, pageSize int) Metadata {
	if totalrecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage: page,
		PageSize:    pageSize,
		FirstPage:   1,
		LastPage:    int(math.Ceil(float64(totalrecords) / float64(pageSize))),
		Total:       totalrecords,
	}
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
