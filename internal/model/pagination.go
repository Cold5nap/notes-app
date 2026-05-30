package model

const MaxPage = 1000

type Pagination struct {
	Page       int
	PerPage    int
	Total      int
	TotalPages int
	Items      []Note
	Query      string
}

func (p Pagination) Offset() int {
	if p.Page < 1 {
		return 0
	}
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
	if p.TotalPages <= 0 {
		return nil
	}

	window := 5
	start := p.Page - window
	if start < 1 {
		start = 1
	}
	end := p.Page + window
	if end > p.TotalPages {
		end = p.TotalPages
	}

	n := end - start + 1
	pages := make([]int, n)
	for i := range pages {
		pages[i] = start + i
	}
	return pages
}
