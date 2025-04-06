-- +goose Up
-- +goose StatementBegin
CREATE TABLE tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	version UNSIGNED INT NOT NULL,
	task_id CHAR(36) NOT NULL,
	title VARCHAR(100) NOT NULL,
	owner VARCHAR(255) NOT NULL,
	description TEXT,
	dueDate DATETIME NOT NULL,
	createdAt DATETIME NOT NULL,
	updatedAt DATETIME NOT NULL,
	status UNSIGNED TINYINT,
	UNIQUE(task_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tasks;
-- +goose StatementEnd
