package repository

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/ayaka/config"
)

type Sqlx struct {
	*sqlx.DB
	Conf *config.Config `inject:"config"`
}

func (s *Sqlx) Startup() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		s.Conf.Database.User,
		s.Conf.Database.Password,
		s.Conf.Database.Host,
		s.Conf.Database.Port,
		s.Conf.Database.DBName,
	)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return err
	}

	s.DB = db

	return nil
}

func (s *Sqlx) Shutdown() error {
	return s.Close()
}
