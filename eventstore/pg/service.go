package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/holdex/hp-backend-lib/eventstore"
	"github.com/holdex/hp-backend-lib/eventstore/pg/pubsub"
	"github.com/holdex/hp-backend-lib/eventstore/switcher"
	"github.com/holdex/hp-backend-lib/log"
)

func NewService(db *sql.DB) libeventstore.Service {
	return &service{db: db, Notifications: pubsub.NewNotifications()}
}

func NewMigrator(db *sql.DB) switcher.Migrator {
	return &service{db: db, Notifications: pubsub.NewNotifications()}
}

type service struct {
	db *sql.DB
	pubsub.Notifications
}

//noinspection SqlResolve
func (s *service) StoreEvents(ctx context.Context, events ...libeventstore.Event) error {
	if len(events) == 0 {
		return nil
	}

	query := `INSERT INTO events (stream_id, stream_type, stream_revision, type, payload, created_at, metadata) VALUES `
	numFields := 7
	var values []interface{}
	for i, ev := range events {
		values = append(values, ev.StreamId, ev.StreamType, ev.StreamRevision, ev.Type, ev.Payload, ev.CreatedAt, ev.Metadata)
		n := i * numFields
		query += `(`
		for j := 0; j < numFields; j++ {
			query += fmt.Sprintf("$%d,", n+j+1)
		}
		query = query[:len(query)-1] + `),`
	}
	query = query[:len(query)-1] + " RETURNING revision"

	rows, err := s.db.QueryContext(ctx, query, values...)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code.Name() == "unique_violation" {
			liblog.Warningf("event stream revision conflicted: %v", err)
			return libeventstore.ErrStreamRevisionAlreadyExists
		}
		return err
	}
	var notifications []pubsub.Notification
	for i := 0; rows.Next(); i++ {
		if err := rows.Scan(&events[i].Revision); err != nil {
			liblog.Errorf("failed to scan returning revision: %v", err)
			break
		}
		notifications = append(notifications, pubsub.Notification{Revision: events[i].Revision, EventType: events[i].Type})
	}
	rows.Close()
	for _, n := range notifications {
		s.Publish(n)
	}
	return nil
}

func (s *service) MigrateEvents(ctx context.Context, events ...libeventstore.StreamEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := `INSERT INTO events (revision, stream_id, stream_type, stream_revision, type, payload, created_at, metadata) VALUES `
	numFields := 8
	var values []interface{}
	for i, ev := range events {
		if ev.Revision == 0 {
			return errors.New("invalid event: no revision given when migrating")
		}
		values = append(values, ev.Revision, ev.StreamId, ev.StreamType, ev.StreamRevision, ev.Type, ev.Payload, ev.CreatedAt, ev.Metadata)
		n := i * numFields
		query += `(`
		for j := 0; j < numFields; j++ {
			query += fmt.Sprintf("$%d,", n+j+1)
		}
		query = query[:len(query)-1] + `),`
	}
	query = query[:len(query)-1]

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, query, values...); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, "SELECT setval('revision', $1)", events[len(events)-1].Revision); err != nil {
		return fmt.Errorf("update revision: %v", err)
	}

	return tx.Commit()
}

//noinspection SqlResolve
func (s *service) LoadEventStream(ctx context.Context, streamID, streamType string, fromStreamRevision uint64) ([]libeventstore.Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT stream_id, stream_type, stream_revision, type, payload, created_at, metadata
FROM events WHERE stream_id = $1 AND stream_type = $2 AND stream_revision > $3 ORDER BY stream_revision  ASC`, streamID, streamType, fromStreamRevision)
	if err != nil {
		return nil, err
	}

	var events []libeventstore.Event
	for rows.Next() {
		var e libeventstore.Event
		if err = rows.Scan(&e.StreamId, &e.StreamType, &e.StreamRevision, &e.Type, &e.Payload, &e.CreatedAt, &e.Metadata); err != nil {
			break
		}
		events = append(events, e)
	}
	rows.Close()
	if err != nil {
		return nil, err
	}
	return events, nil
}

//noinspection SqlResolve
func (s *service) GetRevision(ctx context.Context, eventTypes ...string) (revision uint64, err error) {
	if len(eventTypes) > 0 {
		err = s.db.QueryRowContext(ctx, `SELECT revision FROM events WHERE type = ANY ($1) ORDER BY revision DESC LIMIT 1`, pq.StringArray(eventTypes)).Scan(&revision)
	} else {
		err = s.db.QueryRowContext(ctx, `SELECT revision FROM events ORDER BY revision DESC LIMIT 1`).Scan(&revision)
	}
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

type revision struct {
	r   uint64
	err error
}

func (s *service) StreamEvents(ctx context.Context, fromRevision, buffer uint64, eventTypes ...string) <-chan []libeventstore.StreamEvent {
	startCheckingLastRevision := func(lastSeqNumberCh chan revision) {
		pushRevision := func(r revision) {
			for {
				select {
				case lastSeqNumberCh <- r:
					return
				case <-lastSeqNumberCh:
				}
			}
		}
		if lastRevision, err := s.GetRevision(ctx, eventTypes...); err != nil {
			pushRevision(revision{err: err})
			return
		} else {
			pushRevision(revision{r: lastRevision})
		}

		notificationsCh, unsubscribe := s.Subscribe()
		defer unsubscribe()

		eventTypesMap := make(map[string]bool)
		for _, et := range eventTypes {
			eventTypesMap[et] = true
		}
		for {
			select {
			case <-ctx.Done():
				pushRevision(revision{err: ctx.Err()})
				return
			case lastRevision := <-notificationsCh:
				if _, ok := eventTypesMap[lastRevision.EventType]; ok {
					pushRevision(revision{r: lastRevision.Revision})
				}
			}
		}
	}

	startStreaming := func(stream chan<- []libeventstore.StreamEvent, lastSeqNumberCh <-chan revision) {
		for {
			select {
			case <-ctx.Done():
				stream <- []libeventstore.StreamEvent{{Err: ctx.Err()}}
				return
			case lastSeqNumber := <-lastSeqNumberCh:
				{
					if lastSeqNumber.err != nil {
						stream <- []libeventstore.StreamEvent{{Err: lastSeqNumber.err}}
						close(stream)
						return
					}

					for fromRevision < lastSeqNumber.r {
						select {
						case <-ctx.Done():
							stream <- []libeventstore.StreamEvent{{Err: ctx.Err()}}
							return
						default:
							if evs, err := s.LoadEvents(ctx, fromRevision+1, lastSeqNumber.r, buffer, eventTypes...); err != nil {
								stream <- []libeventstore.StreamEvent{{Err: err}}
								close(stream)
								return
							} else if len(evs) > 0 {
								select {
								case <-ctx.Done():
									stream <- []libeventstore.StreamEvent{{Err: ctx.Err()}}
									return
								case stream <- evs:
									fromRevision = evs[len(evs)-1].Revision
								}
							}
						}
					}
				}
			}
		}
	}

	stream := make(chan []libeventstore.StreamEvent, 10)
	lastSeqNumberCh := make(chan revision, 1)

	go startCheckingLastRevision(lastSeqNumberCh)
	go startStreaming(stream, lastSeqNumberCh)

	return stream
}

//noinspection SqlResolve
func (s *service) LoadEvents(ctx context.Context, fromRevision, toRevision, howMany uint64, eventTypes ...string) ([]libeventstore.StreamEvent, error) {
	var rows *sql.Rows
	var err error
	if len(eventTypes) > 0 {
		rows, err = s.db.QueryContext(ctx, `SELECT revision, stream_id, stream_type, stream_revision, type, payload, created_at, metadata
FROM events WHERE type = ANY ($1) AND revision BETWEEN $2 AND $3 ORDER BY revision LIMIT $4`, pq.StringArray(eventTypes), fromRevision, toRevision, howMany)
	} else {
		rows, err = s.db.QueryContext(ctx, `SELECT revision, stream_id, stream_type, stream_revision, type, payload, created_at, metadata
FROM events WHERE revision BETWEEN $1 AND $2 ORDER BY revision LIMIT $3`, fromRevision, toRevision, howMany)
	}
	if err != nil {
		return nil, err
	}

	var events []libeventstore.StreamEvent
	for rows.Next() {
		var e libeventstore.Event
		if err = rows.Scan(&e.Revision, &e.StreamId, &e.StreamType, &e.StreamRevision, &e.Type, &e.Payload, &e.CreatedAt, &e.Metadata); err != nil {
			break
		}
		events = append(events, libeventstore.StreamEvent{Event: e})
	}
	rows.Close()
	if err != nil {
		return nil, err
	}
	return events, nil
}
