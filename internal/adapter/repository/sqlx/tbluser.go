package sqlx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kataras/golog"
	"github.com/pkg/errors"
	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/tbluser"
	"gitlab.com/ayaka/internal/pkg/customerrors"
)

type TblUserRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblUserRepository) GetByUserCode(ctx context.Context, data *tbluser.Logintbluser) (*tbluser.Logintbluser, error) {
	user := tbluser.Logintbluser{}

	query := "SELECT UserCode, UserName, Pwd FROM tbluser WHERE UserName = ? OR UserCode = ?"

	if err := t.DB.GetContext(ctx, &user, query, data.UserCode, data.UserCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrInvalidUserCode
		}

		return nil, fmt.Errorf("error Login: %w", err)
	}

	return &user, nil
}

func (t *TblUserRepository) GetByEmail(ctx context.Context, email string) (*tbluser.Logintbluser, error) {
	user := tbluser.Logintbluser{}

	query := "SELECT UserCode, UserName FROM tbluser WHERE Email = ?"

	if err := t.DB.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, fmt.Errorf("error get email: %w", err)
	}

	return &user, nil
}

func (t *TblUserRepository) ChangePassword(ctx context.Context, usercode, username, password string) error {
	query := "UPDATE tbluser SET Pwd = ? WHERE UserCode = ? AND UserName = ?"

	if _, err := t.DB.ExecContext(ctx, query, password, usercode, username); err != nil {
		golog.Error(ctx, "Error update password: "+err.Error())
		return err
	}
	return nil
}
