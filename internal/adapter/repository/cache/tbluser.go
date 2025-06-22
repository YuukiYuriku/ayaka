package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/runsystemid/gocache"
	"gitlab.com/ayaka/config"
	"gitlab.com/ayaka/internal/domain/shared/identity"
	"gitlab.com/ayaka/internal/pkg/customerrors"
)

// var keyTemplate = "template-%s"

type TblUserRepository struct {
	Cache  gocache.Service `inject:"cache"`
	Config *config.Config  `inject:"config"`
}

func (t *TblUserRepository) Login(ctx context.Context, accessToken string) (string, error) {
	idToken := identity.NewID().String()

	if err := t.Cache.Put(ctx, idToken, accessToken, time.Duration(t.Config.JWT.JWTDuration)*time.Hour); err != nil {
		fmt.Println("error cache: ", err)
		return "", customerrors.ErrFailedSaveToken
	}
	return idToken, nil
}

func (t *TblUserRepository) GetToken(ctx context.Context, accessToken string) (string, error) {
	var token string
	if err := t.Cache.Get(ctx, accessToken, &token); err != nil {
		log.Printf("Detailed error: %+v", err)
		if err == gocache.ErrNil {
			return "", customerrors.ErrDataNotFound
		}
		return "", err
	}

	return token, nil
}

func (t *TblUserRepository) Logout(ctx context.Context, key string) (int64, error) {
	exists, err := t.Cache.Exists(ctx, key)
	if err != nil {
		return -1, nil
	}
	if !exists {
		return -1, customerrors.ErrDataNotFound
	}

	res, err := t.Cache.Delete(ctx, key)
	if err != nil {
		return -1, nil
	}
	return res, nil
}

func (t *TblUserRepository) SaveEmail(ctx context.Context, email string) error {
	key := fmt.Sprintf("email-%s", email)

	exists, _ := t.Cache.Exists(ctx, key)
	if exists {
		return customerrors.ErrDataAlreadyExists
	}

	if err := t.Cache.Put(ctx, key, email, time.Duration(t.Config.Email.SendDuration)*time.Minute); err != nil {
		return customerrors.ErrFailedSaveToken
	}

	return nil
}

func (t *TblUserRepository) SaveDelTokenForgotPassword(ctx context.Context, token string) (string, error) {
	exists, _ := t.Cache.Exists(ctx, token)
	if exists {
		_, err := t.Cache.Delete(ctx, token)
		return "", err
	} else {
		key := identity.NewID().String()

		if err := t.Cache.Put(ctx, key, token, time.Duration(t.Config.JWT.ChangePassDuration)*time.Minute); err != nil {
			return "", customerrors.ErrFailedSaveToken
		}
		return key, nil
	}
}
