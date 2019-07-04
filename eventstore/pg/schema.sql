CREATE SEQUENCE revision START 2;
CREATE TABLE events
(
	revision        BIGINT DEFAULT nextval('revision') NOT NULL,
	stream_id       VARCHAR                            NOT NULL,
	stream_type     VARCHAR                            NOT NULL,
	stream_revision INTEGER                            NOT NULL,
	type            VARCHAR                            NOT NULL,
	payload         BYTEA                              NOT NULL,
	created_at      BIGINT                             NOT NULL,
	metadata        BYTEA                              NOT NULL
);
CREATE UNIQUE INDEX events_revision_uindex ON events (revision);
CREATE UNIQUE INDEX events_stream_revision_uindex ON events (stream_id, stream_type, stream_revision) WHERE stream_revision > 0;
