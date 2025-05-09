package mysql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/pkg/failure"
)

type ScheduleRepo struct {
	db *sqlx.DB
}

func NewScheduleRepo(db *sqlx.DB) *ScheduleRepo {
	return &ScheduleRepo{
		db: db,
	}
}

func (r *ScheduleRepo) Save(ctx context.Context, schedule *entity.Schedule) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO schedule (user_id, name, end_at, period) VALUES (:user_id, :name, :end_at, :period)", schedule)
	if err != nil {
		return failure.NewInternalError(err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	schedule.Id = value.ScheduleId(id)

	return nil
}

func (r *ScheduleRepo) GetByUser(ctx context.Context, userId value.UserId) ([]*entity.Schedule, error) {
	var schedules []*entity.Schedule
	if err := r.db.SelectContext(ctx, &schedules, "SELECT * FROM schedule WHERE user_id = ?", userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return schedules, failure.NewInternalError(err.Error())
		}
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleRepo) GetById(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*entity.Schedule, error) {
	schedule := new(entity.Schedule)
	if err := r.db.GetContext(ctx, schedule, "SELECT * FROM schedule WHERE user_id = ? AND id = ?", userId, scheduleId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, failure.NewNotFoundError(err.Error())
		}
		return nil, err
	}
	return schedule, nil
}
