package dto

import (
	"encoding/json"
	"fmt"
	"io"
)

func Parse(r io.Reader, payload any) error {
	if r == nil {
		return fmt.Errorf("parsing from nil reader")
	}

	return json.NewDecoder(r).Decode(payload)
}
