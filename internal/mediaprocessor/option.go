package mediaprocessor

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/marketplace/internal/datastore"
	"github.com/videocoin/marketplace/internal/storage"
	"github.com/videocoin/marketplace/pkg/videocoin"
)

type Option func(*MediaProcessor) error

func WithLogger(logger *logrus.Entry) Option {
	return func(mc *MediaProcessor) error {
		mc.logger = logger
		return nil
	}
}

func WithDatastore(ds *datastore.Datastore) Option {
	return func(mc *MediaProcessor) error {
		mc.ds = ds
		return nil
	}
}

func WithStorage(s *storage.Storage) Option {
	return func(mc *MediaProcessor) error {
		mc.storage = s
		return nil
	}
}

func WithVideocoin(vc *videocoin.Client) Option {
	return func(mc *MediaProcessor) error {
		mc.vc = vc
		return nil
	}
}
