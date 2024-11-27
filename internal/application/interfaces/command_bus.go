package interfaces

import "context"

type Command interface{}

type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

type CommandBus interface {
	Dispatch(ctx context.Context, cmd Command) error
}
