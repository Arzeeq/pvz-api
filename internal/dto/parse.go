package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var ErrParsingDTO = errors.New("failed to parse dto")

func Parse(r io.Reader, payload any) error {
	if r == nil {
		return fmt.Errorf("parsing from nil reader")
	}

	err := json.NewDecoder(r).Decode(payload)
	if err != nil {
		return ErrParsingDTO
	}

	return nil
}

func (p *GetPvzParams) FromParams(r *http.Request) error {
	query := r.URL.Query()

	if startDateStr := query.Get("startDate"); startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return err
		}
		p.StartDate = &t
	}

	if endDateStr := query.Get("endDate"); endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return err
		}
		p.EndDate = &t
	}

	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return err
		}
		p.Page = &page
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return err
		}
		p.Limit = &limit
	}

	return nil
}

func CorrectParams(p *GetPvzParams) {
	if p == nil {
		return
	}

	// apply defaults
	if p.Page == nil {
		page := 1
		p.Page = &page
	}
	if p.Limit == nil {
		limit := 10
		p.Limit = &limit
	}
	if p.StartDate == nil {
		startDate := time.Time{}
		p.StartDate = &startDate
	}
	if p.EndDate == nil {
		endDate := time.Now()
		p.EndDate = &endDate
	}

	// check limitations for fields
	limitMin, limitMax := 1, 30
	pageMin := 1

	if *p.Page < pageMin {
		*p.Page = pageMin
	}
	if *p.Limit < limitMin {
		*p.Limit = limitMin
	} else if *p.Limit > limitMax {
		*p.Limit = limitMax
	}
}
