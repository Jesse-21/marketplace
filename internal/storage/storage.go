package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/textileio/go-threads/core/thread"
	bucketsd "github.com/textileio/textile/v2/api/bucketsd/client"
	bucketspb "github.com/textileio/textile/v2/api/bucketsd/pb"
	"github.com/textileio/textile/v2/api/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"time"
)

const (
	Textile    = "textile"
	NftStorage = "nftstorage"
)

type Storage struct {
	backend          string
	textileConfig    *TextileConfig
	nftStorageConfig *NftStorageConfig
	authCtx          context.Context
	ttCli            *bucketsd.Client
	ttRoot           *bucketspb.RootResponse
	nsCli            *NftStorageClient
}

func NewStorage(opts ...Option) (*Storage, error) {
	s := new(Storage)
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	if s.nftStorageConfig != nil {
		s.backend = NftStorage
		s.nsCli = NewNftStorageClient(s.nftStorageConfig.ApiKey)
	}

	if s.textileConfig != nil {
		s.backend = Textile

		dialOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
			grpc.WithPerRPCCredentials(common.Credentials{}),
		}
		cli, err := bucketsd.NewClient("api.hub.textile.io:443", dialOpts...)
		if err != nil {
			return nil, err
		}
		s.ttCli = cli

		authCtx, err := common.CreateAPISigContext(
			common.NewAPIKeyContext(context.Background(), s.textileConfig.AuthKey),
			time.Now().Add(time.Minute),
			s.textileConfig.AuthSecret,
		)
		if err != nil {
			return nil, err
		}

		tid, err := thread.Decode(s.textileConfig.ThreadID)
		if err != nil {
			return nil, fmt.Errorf("invalid textile thread id: %s", err)
		}
		authCtx = common.NewThreadIDContext(authCtx, tid)

		s.authCtx = authCtx

		root, err := s.ttCli.Root(s.authCtx, s.textileConfig.BucketRootKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get bucket root: %s", err)
		}

		s.ttRoot = root
	}

	return s, nil
}

func (s *Storage) PushPath(path string, src io.Reader) (string, error) {
	if s.backend == NftStorage {
		return s.nsCli.Upload(path, src)
	}

	_, _, err := s.ttCli.PushPath(s.authCtx, s.ttRoot.Root.Key, path, src)
	if err != nil {
		return "", err
	}

	links, err := s.ttCli.Links(s.authCtx, s.ttRoot.Root.Key, path)
	if err != nil {
		return "", err
	}

	return links.Ipns, nil
}
