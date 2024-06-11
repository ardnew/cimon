package cimon

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Task func(context.Context, context.CancelFunc) error

type Server[T any] interface {
	Open() error
	Connect() (T, error)
	Respond(T) Task
}

func Serve[T any](ctx context.Context, serv Server[T]) error {
	ctx, cancel := context.WithCancel(ctx)
	group, ctx := errgroup.WithContext(ctx)
	tasks := make(chan Task)
	group.Go(func() error {
		defer close(tasks)
		err := serv.Open()
		if err != nil {
			return err
		}
		cerr := make(chan error)
		defer close(cerr)
		for {
			next := make(chan T)
			go func(serv Server[T], next chan<- T, cerr chan<- error) {
				defer close(next)
				con, err := serv.Connect()
				if err != nil {
					cerr <- err
				}
				next <- con
			}(serv, next, cerr)
			select {
			case conn := <-next:
				select {
				case tasks <- serv.Respond(conn):
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
		group.Go(func() error { return task(ctx, cancel) })
	}
	return group.Wait()
}
