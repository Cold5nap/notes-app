package model

type Pagination struct {
	Page       int
	PerPage    int
	Total      int
	TotalPages int
	Items      []Note
	Query      string
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p Pagination) HasPrev() bool {
	return p.Page > 1
}

func (p Pagination) HasNext() bool {
	return p.Page < p.TotalPages
}

func (p Pagination) PrevPage() int {
	if p.Page <= 1 {
		return 1
	}
	return p.Page - 1
}

func (p Pagination) NextPage() int {
	if p.Page >= p.TotalPages {
		return p.TotalPages
	}
	return p.Page + 1
}

func (p Pagination) Pages() []int {
	n := p.TotalPages
	if n > 10 {
		n = 10
	}
	pages := make([]int, n)
	for i := range pages {
		pages[i] = i + 1
	}
	return pages
}
