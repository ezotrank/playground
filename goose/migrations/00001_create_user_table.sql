-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user"
(
    id   serial NOT NULL,
    name text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "user";
-- +goose StatementEnd
