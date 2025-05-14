package sqlx

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/runsystemid/golog"
	"github.com/stretchr/testify/suite"
	"gitlab.com/ayaka/internal/adapter/repository"
	template "gitlab.com/ayaka/internal/domain/tbluser"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"go.uber.org/mock/gomock"
)

type TblUserRepositorySuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	mockSQL  sqlmock.Sqlmock
	repo     *TblUserRepository
	db       *sqlx.DB
}

func (suite *TblUserRepositorySuite) SetupTest() {
	suite.mockCtrl = gomock.NewController(suite.T())

	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	suite.Require().NoError(err)

	sqlxDb := sqlx.NewDb(mockDb, "mysql")

	golog.Load(golog.Config{})

	suite.mockSQL = mock
	suite.db = sqlxDb
	suite.repo = &TblUserRepository{
		DB: &repository.Sqlx{DB: sqlxDb},
	}
}

func (suite *TblUserRepositorySuite) TearDownTest() {
	suite.db.Close()
	suite.mockCtrl.Finish()
}

// invalid username / usercode
func (suite *TblUserRepositorySuite) TestLogin_ErrNotFound() {
	input := &template.Logintbluser{
		UserCode: "NOTEXISTUSER",
		Password: "password123",
	}

	suite.mockSQL.ExpectQuery("SELECT UserCode, UserName, Pwd FROM tbluser WHERE UserName = ? OR UserCode = ?").
		WithArgs(input.UserCode, input.UserCode).
		WillReturnError(customerrors.ErrDataNotFound)

	result, err := suite.repo.GetByUserCode(context.Background(), input)

	suite.Error(err)
	suite.EqualError(err, "error Login: "+customerrors.ErrDataNotFound.Error())
	suite.Nil(result)
	suite.NoError(suite.mockSQL.ExpectationsWereMet())
}

// valid usercode
func (suite *TblUserRepositorySuite) TestLogin_SuccessWithUserCode() {
	input := &template.Logintbluser{
		UserCode: "USRCD1",
		Password: "Password",
	}

	rows := sqlmock.NewRows([]string{"UserCode", "Pwd"}).
		AddRow(input.UserCode, input.Password)

	suite.mockSQL.ExpectQuery("SELECT UserCode, UserName, Pwd FROM tbluser WHERE UserName = ? OR UserCode = ?").
		WithArgs(input.UserCode, input.UserCode).
		WillReturnRows(rows)

	result, err := suite.repo.GetByUserCode(context.Background(), input)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(input.UserCode, result.UserCode)
	suite.Equal(input.Password, result.Password)
	suite.NoError(suite.mockSQL.ExpectationsWereMet())
}

// valid username
func (suite *TblUserRepositorySuite) TestLogin_SuccessWithUserName() {
	input := &template.Logintbluser{
		UserCode: "USRNM1",
		Password: "Password",
	}

	expectedResult := &template.Logintbluser{
		UserCode: "USER123",
		Password: "password123",
	}

	rows := sqlmock.NewRows([]string{"UserCode", "Pwd"}).
		AddRow(expectedResult.UserCode, expectedResult.Password)

	suite.mockSQL.ExpectQuery("SELECT UserCode, UserName, Pwd FROM tbluser WHERE UserName = ? OR UserCode = ?").
		WithArgs(input.UserCode, input.UserCode).
		WillReturnRows(rows)

	result, err := suite.repo.GetByUserCode(context.Background(), input)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedResult.UserCode, result.UserCode)
	suite.Equal(expectedResult.UserName, result.UserName)
	suite.NoError(suite.mockSQL.ExpectationsWereMet())
}

// Run the test suite
func TestTblUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(TblUserRepositorySuite))
}
