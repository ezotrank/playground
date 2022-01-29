-- +goose Up
-- +goose StatementBegin
ALTER TABLE "user"
    ADD COLUMN "created_at" TIMESTAMP NOT NULL DEFAULT now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "user"
    DROP COLUMN "created_at";
-- +goose StatementEnd
