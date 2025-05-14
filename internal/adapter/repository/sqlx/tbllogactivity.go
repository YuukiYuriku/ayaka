package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/logactivity"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/pkg/customerrors"
)

type TblLogRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblLogRepository) GetActivityLog(ctx context.Context, code, category string) ([]*logactivity.LogActivity, error) {
	table, err := logactivity.TableOf(category)
	if err != nil {
		return nil, err
	}
	primKey, err := logactivity.PrimaryKeyOf(category)
	if err != nil {
		return nil, err
	}

	var detailsActivity []*logactivity.LogActivity

	var create []*logactivity.LogActivity
	query := fmt.Sprintf("SELECT CreateDt AS Date, CONCAT(CreateBy, ' created this data') AS Log FROM %s WHERE %s = ?", table, primKey)
	if err := t.DB.SelectContext(ctx, &create, query, code); err != nil {
		return nil, fmt.Errorf("error log activity: %w", err)
	}
	if create == nil {
		return nil, customerrors.ErrDataNotFound
	}

	detailsActivity = append(detailsActivity, create...)

	query = "SELECT a.LastUpDt AS Date, CONCAT(u.UserName, ' made changes to this data') AS Log FROM tbluser u JOIN tbllogactivity a ON u.UserCode = a.UserCode WHERE a.Code = ? AND a.Category = ? ORDER BY Date DESC"
	var detailsActivityUpdate []*logactivity.LogActivity
	log.Printf("Executing query: %s with code: %s From %s", query, code, table)
	if err := t.DB.SelectContext(ctx, &detailsActivityUpdate, query, code, category); err != nil {
		log.Printf("Detailed error: %+v", err)
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error log activity: %w", err)
		}
	}

	if detailsActivityUpdate != nil {
		detailsActivity = append(detailsActivityUpdate, detailsActivity...)
	}

	for _, detail := range detailsActivity {
		detail.Date = share.FormatDate(detail.Date)
	}

	return detailsActivity, nil
}
