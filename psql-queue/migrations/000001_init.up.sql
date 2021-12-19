CREATE TYPE event_status AS ENUM ('new', 'running', 'success', 'error');

CREATE TABLE events
(
    id         SERIAL,
    key        UUID,
    status     event_status,
    error      text,
    created_at timestamp,
    updated_at timestamp
);

CREATE UNIQUE INDEX name ON events (key);

CREATE OR REPLACE FUNCTION events_status_notify()
    RETURNS trigger AS
$$
BEGIN
    PERFORM pg_notify('events_status_channel', NEW.status::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER events_status
    AFTER INSERT OR UPDATE OF status
    ON events
    FOR EACH ROW
EXECUTE PROCEDURE events_status_notify();