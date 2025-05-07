-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE schedule (
    id      int auto_increment primary key,
    user_id bigint       not null,
    name    varchar(255) not null,
    end_at  date         null,
    period  bigint       not null
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE schedule;

-- +goose StatementEnd
