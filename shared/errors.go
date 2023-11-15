package shared

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrNoRemoteActions = errors.New("service actions not available")
)

type ErrAccountNotExists struct {
	Prefix,
	ID string
}

func (e ErrAccountNotExists) Error() string {
	return fmt.Sprintf("%s: account not exists (id: %s)", e.Prefix, e.ID)
}

func NewErrSnapshotNotFound(prefix string, id string) ErrSnapshotNotFound {
	return ErrSnapshotNotFound{
		Prefix: prefix,
		ID:     id,
	}
}

type ErrSnapshotNotFound struct {
	Prefix,
	ID string
}

func (e ErrSnapshotNotFound) Error() string {
	return fmt.Sprintf("%s: snapshot not exists (id: %s)", e.Prefix, e.ID)
}
