package shared

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrNoRemoteActions = errors.New("remote actions not available")
)

func NewErrRemoteNotFound(name RemoteName) ErrRemoteNotFound {
	return ErrRemoteNotFound{
		Name: name,
	}
}

type ErrRemoteNotFound struct {
	Name RemoteName
}

func (e ErrRemoteNotFound) Error() string {
	return fmt.Sprintf("remote '%s' not found", e.Name.String())
}

func NewErrNoAvailableRemotes() ErrNoAvailableRemotes {
	return ErrNoAvailableRemotes{}
}

type ErrNoAvailableRemotes struct {
}

func (e ErrNoAvailableRemotes) Error() string {
	return "No available remotes"
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
