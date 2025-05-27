package service

import (
	"context"
	"fmt"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/config"
	"gitlab.com/ayaka/internal/domain/tbluser"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/custommiddleware"
	"gitlab.com/ayaka/internal/pkg/email"
	share "gitlab.com/ayaka/internal/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type TblUserService interface {
	Login(ctx context.Context, data *tbluser.Logintbluser) (string, int, error)
	Logout(ctx context.Context, key string) (int64, error)
	SendEmailForgotPassword(ctx context.Context, emailUser string) (*tbluser.Logintbluser, error)
	ChangePassword(ctx context.Context, token string, user *tbluser.ForgotPassword) error
}

type TblUser struct {
	TemplateRepo      tbluser.Repository                   `inject:"tblUserRepository"`
	TemplateCacheRepo tbluser.RepositoryLoginCache         `inject:"tblUserCacheRepository"`
	JwtHandler        *share.JwtHandler                    `inject:"jwtHandler"`
	EmailHandler      *email.MailSMTP                      `inject:"mail"`
	Conf              *config.Config                       `inject:"config"`
	Log               *custommiddleware.LogActivityHandler `inject:"logActivity"`
}

func (t *TblUser) Login(ctx context.Context, data *tbluser.Logintbluser) (string, int, error) {
	user, err := t.TemplateRepo.GetByUserCode(ctx, data)

	if err != nil {
		golog.Error(ctx, "Error login: "+err.Error(), err)
		return "", 0, err
	}
	fmt.Println("pw: ", user.Password)
	fmt.Println("pw: ", data.Password)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		golog.Error(ctx, "Error compare password: "+err.Error(), err)
		return "", 0, customerrors.ErrInvalidPassword
	}

	token, _ := t.JwtHandler.GenerateToken(user.UserCode, user.UserName, t.Conf.JWT.JWTKEY, time.Duration(t.Conf.JWT.JWTDuration), time.Hour)
	tokenReturn, err := t.TemplateCacheRepo.Login(ctx, token)

	if err != nil {
		return "", 0, nil
	}

	return tokenReturn, t.Conf.JWT.JWTDuration * 3600, nil
}

func (t *TblUser) Logout(ctx context.Context, key string) (int64, error) {
	res, err := t.TemplateCacheRepo.Logout(ctx, key)

	if err != nil {
		golog.Error(ctx, "Error logout:; "+err.Error(), err)
		return res, err
	}

	return res, nil
}

func (t *TblUser) SendEmailForgotPassword(ctx context.Context, emailUser string) (*tbluser.Logintbluser, error) {
	user, err := t.TemplateRepo.GetByEmail(ctx, emailUser)
	if err != nil {
		golog.Error(ctx, "Error get user by email: "+err.Error(), err)
		return nil, err
	}

	err = t.TemplateCacheRepo.SaveEmail(ctx, emailUser)
	if err != nil {
		golog.Error(ctx, "Error save email: "+err.Error(), err)
		return nil, err
	}

	tokenJWT, err := t.JwtHandler.GenerateToken(user.UserCode, user.UserName, t.Conf.JWT.ChangePassKey, time.Duration(t.Conf.JWT.ChangePassDuration), time.Minute)
	if err != nil {
		golog.Error(ctx, "Error send email: "+err.Error(), err)
		return nil, err
	}

	token, err := t.TemplateCacheRepo.SaveDelTokenForgotPassword(ctx, tokenJWT)
	if err != nil {
		return nil, err
	}
	body := fmt.Sprintf(`
			<body>
				<h1>Hi %s,</h1>
				<p>Click this <a href="%s%s%s">Link</a> to change your R1 password</p>
			</body>
		`, user.UserName, t.Conf.Domain.FrontendDomain, t.Conf.Domain.ForgotPass, token)

	p := &email.EmailSend{
		EmailFrom: t.Conf.Email.EmailFrom,
		EmailTo:   emailUser,
		EmailSubj: "Change Password R1",
		EmailBody: body,
	}

	go t.EmailHandler.Send(ctx, p)
	go t.Log.LogUserInfo(emailUser, "INFO", "Send Email")
	// if err != nil {
	// 	golog.Error(ctx, "Error send email: "+err.Error(), err)
	// }
	return user, nil
}

func (t *TblUser) ChangePassword(ctx context.Context, token string, user *tbluser.ForgotPassword) error {
	getToken, err := t.TemplateCacheRepo.GetToken(ctx, token)
	if err != nil {
		return err
	}

	claims, err := t.JwtHandler.DecodeAndVerifyJWT(getToken, []byte(t.Conf.JWT.ChangePassKey))
	if err != nil {
		return err
	}

	tempUser := &tbluser.Logintbluser{
		UserName: claims.UserName,
		UserCode: claims.UserCode,
	}

	result, err := t.TemplateRepo.GetByUserCode(ctx, tempUser)

	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password)); err == nil {
		return customerrors.ErrInvalidPassword
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	if err = t.TemplateRepo.ChangePassword(ctx, claims.UserCode, claims.UserName, string(hashPassword)); err != nil {
		return err
	}

	_, err = t.TemplateCacheRepo.SaveDelTokenForgotPassword(ctx, token)
	if err != nil {
		return err
	}

	go t.Log.LogUserInfo(tempUser.UserCode, "INFO", "Change Password")
	return nil
}
