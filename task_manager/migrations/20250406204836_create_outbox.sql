-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox (
    sequence int auto_increment primary key,
    event_id char(36) not null,
    type varchar(255) not null,
    version unsigned int not null,
    aggregate_id char(36) not null,
    payload text,
	metadata text,
    timestamp datetime not null,
    status int default 1,
    reserved_by text,
    reserved_at timestamp,
    attempt int default 0,
    error text,
    unique (event_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE outbox;

-- +goose StatementEnd
