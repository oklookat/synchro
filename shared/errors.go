package shared

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrNoRemoteActions = errors.New("remote actions not available")
)

func NewErrRemoteNotFound(prefix string, name RemoteName) ErrRemoteNotFound {
	return ErrRemoteNotFound{
		Prefix: prefix,
		Name:   name,
	}
}

type ErrRemoteNotFound struct {
	Prefix string
	Name   RemoteName
}

func (e ErrRemoteNotFound) Error() string {
	return fmt.Sprintf("%s: remote with name '%s' not found", e.Prefix, e.Name.String())
}

func NewErrNoAvailableRemotes(prefix string) ErrNoAvailableRemotes {
	return ErrNoAvailableRemotes{
		Prefix: prefix,
	}
}

type ErrNoAvailableRemotes struct {
	Prefix string
}

func (e ErrNoAvailableRemotes) Error() string {
	return fmt.Sprintf("%s: no available remotes", e.Prefix)
}

func NewErrAccountNotExists(prefix string, id string) ErrAccountNotExists {
	return ErrAccountNotExists{
		Prefix: prefix,
		ID:     id,
	}
}

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
