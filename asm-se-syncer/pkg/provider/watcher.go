package provider

import (
	"context"
)

// Watcher is the interface of each provider
type Watcher interface {
	Run(ctx context.Context)
	Cache() Cache
	Prefix() string
	ToNamespace() string
	WatcherType() string
}
