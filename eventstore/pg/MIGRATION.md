# Migration steps for eventstore to Google Cloud SQL Postgres

## Setup database, user and permissions

### Set envs
```sql
\set database_name my_database
\set user_name my_user
```

### Create a database and connect to it
```sql
CREATE DATABASE :database_name; 
\c :database_name
```

### Create the schema
```sql
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
``` 

### Create a user
```sql
CREATE USER :user_name; 
\password :user_name
```

### Grants permissions
```sql
GRANT CONNECT ON DATABASE :database_name TO :user_name;
GRANT SELECT,INSERT ON events TO :user_name;
GRANT USAGE ON SEQUENCE revision TO :user_name;
```

> **NOTE**: if new database will me used for migration from another database then the user needs additional permission on the revision.
> The permission should be removed after migration finished

Add permission
```sql
GRANT UPDATE ON SEQUENCE revision TO :user_name;
```
Remove permission
```sql
REVOKE UPDATE ON SEQUENCE revision FROM :user_name; 
```
