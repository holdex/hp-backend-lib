package libeventstore

import (
	"context"
	"errors"
)

var (
	ErrStreamRevisionAlreadyExists = errors.New("stream revision already exists")
)

type Service interface {
	StoreEvents(ctx context.Context, events ...Event) error
	LoadEventStream(ctx context.Context, streamID, streamType string, fromStreamRevision uint64) ([]Event, error)
	GetRevision(ctx context.Context, eventTypes ...string) (uint64, error)
	StreamEvents(ctx context.Context, fromRevision, buffer uint64, eventTypes ...string) <-chan StreamEvent
	LoadEvents(ctx context.Context, fromRevision, toRevision, howMany uint64, eventTypes ...string) ([]Event, error)
}

type Event struct {
	Revision       uint64
	StreamId       string
	StreamType     string
	StreamRevision uint64
	Type           string
	Payload        []byte
	CreatedAt      int64
	Metadata       []byte
}

type StreamEvent struct {
	Event
	Err error
}
