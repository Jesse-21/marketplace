package datastore

import (
	"context"
	"github.com/gocraft/dbr/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	_ "github.com/lib/pq" // nolint
	"github.com/videocoin/marketplace/internal/model"
	"github.com/videocoin/marketplace/pkg/dbrutil"
)

type Datastore struct {
	conn *dbr.Connection

	Accounts  *AccountDatastore
	Assets    *AssetDatastore
	Tokens    *TokenDatastore
	Orders    *OrderDatastore
	ChainMeta *ChainMetaDatastore
}

func NewDatastore(ctx context.Context, uri string) (*Datastore, error) {
	ds := new(Datastore)

	logger := dbrutil.NewLogrusLogger(ctxlogrus.Extract(ctx))

	conn, err := dbr.Open("postgres", uri, logger)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	ds.conn = conn

	accountsDs, err := NewAccountDatastore(ctx, conn)
	if err != nil {
		return nil, err
	}

	ds.Accounts = accountsDs

	assetsDs, err := NewAssetDatastore(ctx, conn)
	if err != nil {
		return nil, err
	}

	ds.Assets = assetsDs

	tokensDs, err := NewTokenDatastore(ctx, conn)
	if err != nil {
		return nil, err
	}

	ds.Tokens = tokensDs

	ordersDs, err := NewOrderDatastore(ctx, conn)
	if err != nil {
		return nil, err
	}

	ds.Orders = ordersDs

	chainMetaDs, err := NewChainMetaDatastore(ctx, conn)
	if err != nil {
		return nil, err
	}

	ds.ChainMeta = chainMetaDs

	return ds, nil
}

func (ds *Datastore) GetAssetsList(ctx context.Context, fltr *AssetsFilter, opts *LimitOpts) ([]*model.Asset, error) {
	accounts, err := ds.Accounts.List(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	assets, err := ds.Assets.List(ctx, fltr, opts)
	if err != nil {
		return nil, err
	}

	JoinAccountsToAsset(ctx, assets, accounts)

	return assets, nil
}

func (ds *Datastore) GetAssetsListCount(ctx context.Context, fltr *AssetsFilter) (int64, error) {
	count, err := ds.Assets.Count(ctx, fltr)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func JoinAccountsToAsset(ctx context.Context, assets []*model.Asset, accounts []*model.Account) {
	byID := map[int64]*model.Account{}
	for _, item := range accounts {
		byID[item.ID] = item
	}
	for _, asset := range assets {
		asset.Account = byID[asset.CreatedByID]
	}
}
