-- +goose Up
-- +goose StatementBegin
create table backlogs (
	owner varchar(255),
	version unsigned integer default 1,
	UNIQUE(owner)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table backlogs;
-- +goose StatementEnd
