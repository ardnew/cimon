package server

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Task func(context.Context, context.CancelFunc) error

type Server[T any] interface {
	Open(ctx context.Context) error
	Connect(ctx context.Context, clients chan<- T) error
	Respond(ctx context.Context, client T) Task
}

func Run[T any](ctx context.Context, cancel context.CancelFunc, serv Server[T]) error {
	group, ctx := errgroup.WithContext(ctx)
	tasks := make(chan Task)
	group.Go(func() error {
		defer close(tasks)
		err := serv.Open(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		cerr := make(chan error)
		defer close(cerr)
		for {
			next := make(chan T)
			go func(serv Server[T], next chan<- T, cerr chan<- error) {
				err := serv.Connect(ctx, next)
				if err != nil {
					cerr <- err
				}
			}(serv, next, cerr)
			select {
			case conn := <-next:
				select {
				case tasks <- serv.Respond(ctx, conn):
				case <-ctx.Done():
					return nil
				}
			case err := <-cerr:
				return err
			case <-ctx.Done():
				return nil
			}
		}
	})
	for task := range tasks {
		task := task //nolint:copyloopvar
		group.Go(func() error { return task(ctx, cancel) })
	}
	return errors.WithStack(group.Wait())
}
