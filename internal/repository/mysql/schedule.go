package mysql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"schedule/internal/entity"
)

type ScheduleRepo struct {
	db *sqlx.DB
}

func NewScheduleRepo(db *sqlx.DB) *ScheduleRepo {
	return &ScheduleRepo{
		db: db,
	}
}

func (r *ScheduleRepo) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS schedule (
		id      int auto_increment primary key,
		user_id bigint       not null,
		name    varchar(255) not null,
		end_at  date         null,
		period  bigint       not null
	);
	`
	if _, err := r.db.Exec(query); err != nil {
		return err
	}
	return nil
}

func (r *ScheduleRepo) Save(ctx context.Context, schedule *entity.Schedule) error {
	res, err := r.db.NamedExecContext(ctx, "INSERT INTO schedule (user_id, name, end_at, period) VALUES (:user_id, :name, :end_at, :period)", schedule)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	schedule.Id = int(id)

	return nil
}

func (r *ScheduleRepo) GetByUser(ctx context.Context, userId int64) ([]entity.Schedule, error) {
	var schedules []entity.Schedule
	if err := r.db.SelectContext(ctx, &schedules, "SELECT * FROM schedule WHERE user_id = ?", userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return schedules, nil
		}
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleRepo) GetById(ctx context.Context, userId int64, scheduleId int) (*entity.Schedule, error) {
	schedule := new(entity.Schedule)
	if err := r.db.GetContext(ctx, schedule, "SELECT * FROM schedule WHERE user_id = ? AND id = ?", userId, scheduleId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return schedule, nil
		}
		return nil, err
	}
	return schedule, nil
}
