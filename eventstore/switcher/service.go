package switcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/holdex/hp-backend-lib/eventstore"
	"github.com/holdex/hp-backend-lib/log"
)

var ErrDataSourceNotAvailable = errors.New("data source not available")

type Migrator interface {
	libeventstore.Service
	MigrateEvents(ctx context.Context, events ...libeventstore.Event) error
}

func NewService(ctx context.Context, s1 libeventstore.Service, s2 Migrator) libeventstore.Service {
	s := &service{s1: s1, s2: s2, cancelSubscriptions: make(chan struct{})}
	go func() {
		for {
			if err := func() error {
				s2revision, err := s.s2.GetRevision(context.Background())
				if err != nil {
					s.Lock()
					s.activeSvc = 1
					s.Unlock()
					return fmt.Errorf("s2.GetRevision: %v", err)
				}

				s1revision, err := s.s1.GetRevision(context.Background())
				if err != nil {
					return fmt.Errorf("s1.GetRevision: %v", err)
				}

				if s2revision > s1revision {
					liblog.Info("New datasource is ahead. All synced")
					s.Lock()
					s.activeSvc = 2
					close(s.cancelSubscriptions)
					s.Unlock()
					return nil
				}
				s.Lock()
				s.activeSvc = 1
				s.Unlock()

				return s.sync(ctx, s1revision, s2revision)
			}(); err != nil {
				liblog.Errorf("failed: %v", err)
				time.Sleep(1 * time.Second)
			} else {
				liblog.Info("New datasource synced and active")
				return
			}
		}
	}()
	return s
}

func (s *service) sync(ctx context.Context, s1revision, s2revision uint64) error {
	buffer := 1000
	isLocked := false
	defer func() {
		if isLocked {
			s.Unlock()
		}
	}()
	for {
		if ok, err := func() (bool, error) {
			events, err := s.s1.LoadEvents(ctx, s2revision+1, s1revision, uint64(buffer))
			if err != nil {
				return false, err
			}

			if len(events) < buffer { // Got less than buffer events, we are near the end

				if isLocked { // If locked, transfer remaining events and switch traffic
					if err := s.s2.MigrateEvents(ctx, events...); err != nil {
						return false, err
					}
					s.activeSvc = 2
					close(s.cancelSubscriptions)
					return true, nil
				}
				// If not locked, lock and sync remaining
				s.Lock()
				isLocked = true

			} else if isLocked {

				// Got more than buffer events, we are moving away form the end
				return false, errors.New("could not succeed the sync and switch, s1 got ahead by more than 100 events")

			} else {

				// Transfer events
				if err := s.s2.MigrateEvents(ctx, events...); err != nil {
					return false, err
				}
				if len(events) > 0 {
					s2revision = events[len(events)-1].Revision
				}
			}

			s1revision, err = s.s1.GetRevision(context.Background())
			if err != nil {
				return false, fmt.Errorf("s1.1.GetRevision: %v", err)
			}
			return false, nil
		}(); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
}

type service struct {
	s1                  libeventstore.Service
	s2                  Migrator
	activeSvc           int
	cancelSubscriptions chan struct{}
	sync.RWMutex
}

func (s *service) GetRevision(ctx context.Context, eventTypes ...string) (uint64, error) {
	s.RLock()
	defer s.RUnlock()

	switch s.activeSvc {
	case 1:
		return s.s1.GetRevision(ctx, eventTypes...)
	case 2:
		return s.s2.GetRevision(ctx, eventTypes...)
	default:
		return 0, ErrDataSourceNotAvailable
	}
}

func (s *service) StoreEvents(ctx context.Context, events ...libeventstore.Event) error {
	s.RLock()
	defer s.RUnlock()

	switch s.activeSvc {
	case 1:
		return s.s1.StoreEvents(ctx, events...)
	case 2:
		return s.s2.StoreEvents(ctx, events...)
	default:
		return ErrDataSourceNotAvailable
	}
}

func (s *service) LoadEventStream(ctx context.Context, streamID, streamType string, fromStreamRevision uint64) ([]libeventstore.Event, error) {
	s.RLock()
	defer s.RUnlock()

	switch s.activeSvc {
	case 1:
		return s.s1.LoadEventStream(ctx, streamID, streamType, fromStreamRevision)
	case 2:
		return s.s2.LoadEventStream(ctx, streamID, streamType, fromStreamRevision)
	default:
		return nil, ErrDataSourceNotAvailable
	}
}

func (s *service) StreamEvents(ctx context.Context, fromRevision, buffer uint64, eventTypes ...string) <-chan libeventstore.StreamEvent {
	s.RLock()
	defer s.RUnlock()

	switch s.activeSvc {
	case 1:
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			<-s.cancelSubscriptions
			cancel()
		}()
		return s.s1.StreamEvents(ctx, fromRevision, buffer, eventTypes...)
	case 2:
		return s.s2.StreamEvents(ctx, fromRevision, buffer, eventTypes...)
	default:
		ch := make(chan libeventstore.StreamEvent, 1)
		ch <- libeventstore.StreamEvent{Err: ErrDataSourceNotAvailable}
		return ch
	}
}

func (s *service) LoadEvents(ctx context.Context, fromRevision, toRevision, howMany uint64, eventTypes ...string) ([]libeventstore.Event, error) {
	s.RLock()
	defer s.RUnlock()

	switch s.activeSvc {
	case 1:
		return s.s1.LoadEvents(ctx, fromRevision, toRevision, howMany, eventTypes...)
	case 2:
		return s.s1.LoadEvents(ctx, fromRevision, toRevision, howMany, eventTypes...)
	default:
		return nil, ErrDataSourceNotAvailable
	}
}
