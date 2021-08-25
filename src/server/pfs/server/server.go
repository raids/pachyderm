package server

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (pfsserver.APIServer, error) {
	env, err := EnvFromServiceEnv(senv, txnEnv)
	if err != nil {
		return nil, err
	}
	env.EtcdPrefix = etcdPrefix
	a, err := newAPIServer(*env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}

// TODO: This isn't working yet, but this is the goal
func NewTestServer(t testing.TB) pfs.APIServer {
	db := dockertestenv.NewTestDB(t)
	ctx := context.Background()
	etcdEnv := testetcd.NewEnv(t)
	err := dbutil.WithTx(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CREATE SCHEMA storage`)
		require.NoError(t, err)
		_, err = tx.Exec(`CREATE SCHEMA pfs`)
		require.NoError(t, err)
		_, err = tx.Exec(`CREATE SCHEMA collections`)
		require.NoError(t, err)
		require.NoError(t, fileset.SetupPostgresStoreV0(tx))
		require.NoError(t, chunk.SetupPostgresStoreV0(tx))
		require.NoError(t, collection.SetupPostgresCollections(ctx, tx))
		require.NoError(t, collection.SetupPostgresV0(ctx, tx))
		return nil
	})
	require.NoError(t, err)
	oc, _ := obj.NewTestClient(t)
	srv, err := newAPIServer(Env{
		BackgroundContext: ctx,
		ObjectClient:      oc,
		DB:                db,
		Logger:            logrus.StandardLogger(),
		AuthServer:        nil,
		PPSServer:         nil,
		StorageConfig: serviceenv.StorageConfiguration{
			StorageCompactionMaxFanIn: 2,
		},
		EtcdClient: etcdEnv.EtcdClient,
		TxnEnv:     nil,
	})
	require.NoError(t, err)
	return srv
}

func NewTestClient(t testing.TB) pfs.APIClient {
	srv := NewTestServer(t)
	gc := grpcutil.NewTestClient(t, func(gs *grpc.Server) {
		pfs.RegisterAPIServer(gs, srv)
	})
	return pfs.NewAPIClient(gc)
}
