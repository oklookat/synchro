package shared

import (
	"errors"
	"fmt"

	"github.com/oklookat/synchro/streaming"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type ErrAccountNotExists struct {
	Prefix,
	ID string
}

func (e ErrAccountNotExists) Error() string {
	return fmt.Sprintf("%s: account not exists (id: %s)", e.Prefix, e.ID)
}

func NewErrServiceNotFound(name streaming.ServiceName) ErrServiceNotFound {
	return ErrServiceNotFound{
		Name: name,
	}
}

type ErrServiceNotFound struct {
	Name streaming.ServiceName
}

func (e ErrServiceNotFound) Error() string {
	return fmt.Sprintf("service '%s' not found", e.Name)
}
