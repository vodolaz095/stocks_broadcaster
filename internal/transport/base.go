package transport

import "context"

type Transport interface {
	Name() string
	Ping(context.Context) error
	Close(context.Context) error
}
