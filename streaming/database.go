package streaming

import (
	"context"
	"strconv"
	"time"
)

// Example: artist entity ID from DB.
type DatabaseEntityID uint64

func (e DatabaseEntityID) String() string {
	return strconv.FormatUint(uint64(e), 10)
}

func (e *DatabaseEntityID) FromString(val string) error {
	conv, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return err
	}
	*e = DatabaseEntityID(conv)
	return err
}

type (
	// Streaming service in database.
	Database interface {
		// Unique ID.
		ID() string

		// Streaming service name.
		Name() ServiceName

		// Alias: any text specified by the user so that he can distinguish one account from another.
		//
		// Auth: any auth like json tokens.
		//
		// After creating the account, the AssignActions method will be called in the streaming.
		CreateAccount(alias string, auth string) (Account, error)

		// All accounts in streaming.
		Accounts(context.Context) ([]Account, error)

		// Get account by ID. Returns nil, nil if account not found.
		Account(id string) (Account, error)

		// Streaming service actions.
		Actions() (ServiceActions, error)
	}

	// Account.
	Account interface {
		// Unique ID.
		ID() string

		// Database that account belongs to.
		Database() (Database, error)

		// Streaming service name.
		ServiceName() ServiceName

		// Any text specified by the user so that he can distinguish one account from another.
		Alias() string

		// Set account alias.
		SetAlias(string) error

		// Account auth data like json tokens.
		Auth() string

		// Set account auth.
		SetAuth(string) error

		// Actions for account.
		Actions() (AccountActions, error)

		// Time when account was added.
		AddedAt() time.Time

		// Delete account from database.
		Delete() error
	}
)
