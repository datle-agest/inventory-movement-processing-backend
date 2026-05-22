package core

type Pagination struct {
	Page  int `json:"page" example:"1"`
	Limit int `json:"limit" example:"10"`
	Total int `json:"total" example:"100"`
}

func (p *Pagination) Process() {
	if p.Page < 1 {
		p.Page = 1
	}

	if p.Limit <= 0 {
		p.Limit = 10
	}

	if p.Limit >= 200 {
		p.Limit = 200
	}
}
