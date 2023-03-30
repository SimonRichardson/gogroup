package gogroup_test

import (
	"context"
	"testing"
	"time"

	"github.com/SimonRichardson/gogroup"
)

const (
	shortTime = 10 * time.Millisecond
	longTime  = 10 * time.Second
)

func TestEmptyGroup(t *testing.T) {
	group := gogroup.New(context.Background())
	if err := group.Err(); err != nil {
		t.Errorf("Expected nil, recieved %v", err)
	}

	group.Cancel()

	checkCleanKill(t, group)
}

func TestGroup(t *testing.T) {
	group := gogroup.New(context.Background())
	group.Go(func(ctx context.Context) error {
		return nil
	})

	// No need to cancel, the group will cancel itself.
	checkCleanKill(t, group)
}

func TestGroupContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	group := gogroup.New(ctx)
	group.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
		case <-time.After(shortTime):
			t.Errorf("Expected context to be cancelled")
		}
		return nil
	})

	cancel()

	checkCleanKill(t, group)
}

func checkCleanKill(t *testing.T, group *gogroup.Group) {
	if err := group.Wait(); err != nil {
		t.Errorf("Expected nil, recieved %v", err)
	}
	if err := group.Err(); err != nil {
		t.Errorf("Expected nil, recieved %v", err)
	}

	select {
	case <-group.Dying():
	default:
		t.Errorf("Expected group to be dying")
	}

	select {
	case <-group.Done():
	default:
		t.Errorf("Expected group to be done")
	}
}
