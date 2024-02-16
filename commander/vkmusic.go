package commander

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/remote/vkmusic"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/vkmauth"
)

type (
	VkMusicOnCode interface {
		Got(currentMethod, resendMethod string) VkMusicOnCodeGetter
	}

	VkMusicOnCodeGetter interface {
		Code() (string, error)
		Resend() bool
	}
)

func NewVkMusic() *VkMusic {
	return &VkMusic{}
}

type VkMusic struct {
}

func (e VkMusic) RemoteName() string {
	return vkmusic.RemoteName.String()
}

func (e VkMusic) NewAccount(
	alias,
	phone string,
	password string,
	onCode VkMusicOnCode,
	deadlineSeconds int,
) (string, error) {

	if onCode == nil {
		return "", errors.New("vkmusic: nil onCode")
	}
	var (
		account shared.Account
		err     error
	)
	if err := execTask(deadlineSeconds, func(ctx context.Context) error {
		onCoded := func(by vkmauth.CodeSended) (vkmauth.GotCode, error) {
			getr := onCode.Got(by.Current.String(), by.Resend.String())
			coded, err := getr.Code()
			got := vkmauth.GotCode{
				Code:   coded,
				Resend: getr.Resend(),
			}
			return got, err
		}
		account, err = vkmusic.NewAccount(ctx, &alias, phone, password, onCoded)
		return err
	}); err != nil {
		return "", err
	}

	return account.ID(), err
}

func (e VkMusic) Reauth(
	accountId,
	phone,
	password string,
	onCode VkMusicOnCode,
	deadlineSeconds int,
) error {
	return execTask(deadlineSeconds, func(ctx context.Context) error {
		accById, err := accountByID(accountId)
		if err != nil {
			return err
		}

		return vkmusic.Reauth(
			ctx, accById, phone, password,
			func(by vkmauth.CodeSended) (vkmauth.GotCode, error) {
				getr := onCode.Got(by.Current.String(), by.Resend.String())
				coded, err := getr.Code()
				got := vkmauth.GotCode{
					Code:   coded,
					Resend: getr.Resend(),
				}
				return got, err
			},
		)
	})
}
