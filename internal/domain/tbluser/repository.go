package tbluser

import (
	"context"
)

type Repository interface {
	//login
	GetByUserCode(ctx context.Context, data *Logintbluser) (*Logintbluser, error)
	// get user by email
	GetByEmail(ctx context.Context, email string) (*Logintbluser, error)
	// change password
	ChangePassword(ctx context.Context, usercode, username, password string) error
}


type RepositoryLoginCache interface {
	//save token
	Login(ctx context.Context, accessToken string) (string, error)
	//get token
	GetToken(ctx context.Context, accessToken string) (string, error)
	// logout -> delete token from redis
	Logout(ctx context.Context, key string) (int64, error)
	// Save email user 
	SaveEmail(ctx context.Context, email string) error
	// Save or delete token for forgot password
	SaveDelTokenForgotPassword(ctx context.Context, token string) (string, error)
}