package gogroup

import (
	"context"
	"errors"
	"sync"
)

type GroupFunc func(context.Context) error

type Group struct {
	mutex  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	dead   chan struct{}
	alive  int
	reason error
}

// New creates a new Group.
func New(root context.Context) *Group {
	ctx, cancel := context.WithCancel(root)
	return &Group{
		ctx:    ctx,
		cancel: cancel,

		dead: make(chan struct{}),
	}
}

func (g *Group) Go(f GroupFunc) {
	g.mutex.Lock()
	g.alive++
	g.mutex.Unlock()

	go g.run(f)
}

func (g *Group) Wait() error {
	<-g.dead

	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.reason
}

func (g *Group) Cancel() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.kill(nil)
}

func (g *Group) Dying() <-chan struct{} {
	return g.ctx.Done()
}

func (g *Group) Done() <-chan struct{} {
	return g.dead
}

func (g *Group) Err() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.reason
}

func (g *Group) run(f GroupFunc) {
	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()

	err := f(ctx)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.alive--
	if g.alive == 0 || err != nil {
		g.kill(err)
	}
}

func (g *Group) kill(err error) {
	if g.reason != nil {
		return
	}

	g.reason = err
	g.cancel()
	if g.alive == 0 {
		close(g.dead)
	}

	if g.reason == nil {
		ctxErr := g.ctx.Err()
		if !errors.Is(ctxErr, context.Canceled) {
			g.reason = ctxErr
		}
	}
}
