package handler

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	QQueryErr            = errors.New("query error")
	FFormatResponseError = errors.New("format response error")
)

//

type ErrNotFound struct {
	id uuid.UUID
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("id %s not found", e.id)
}

//

type ErrInvalidArg struct {
	id   uuid.UUID
	body []byte
}

func (e ErrInvalidArg) Error() string {
	if e.id != uuid.Nil {
		return fmt.Sprintf("id %s is invalid", e.id)
	}
	if len(e.body) != 0 {
		return fmt.Sprintf("body %s is invalid", string(e.body))
	}

	return "invalid args"
}
