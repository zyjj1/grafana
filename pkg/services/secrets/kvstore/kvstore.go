package kvstore

import (
	"context"
	"flag"
	"fmt"

	"github.com/grafana/grafana/pkg/infra/log"
	pb "github.com/grafana/grafana/pkg/plugins/backendplugin/secretsmanagerplugin"
	"github.com/grafana/grafana/pkg/services/secrets"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// Wildcard to query all organizations
	AllOrganizations = -1
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func ProvideService(sqlStore sqlstore.Store, secretsService secrets.Service, cfg *setting.Cfg) SecretsKVStore {
	logger := log.New("secrets.kvstore")
	usePlugin := cfg.SectionWithEnvOverrides("secrets").Key("use_plugin").MustBool()
	fmt.Print("\n\n\n", usePlugin, "\n\n\n")
	if usePlugin {
		conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Error("did not connect: %v", err)
		}
		// defer conn.Close()
		c := pb.NewRemoteSecretsManagerClient(conn)

		return &secretsKVStorePlugin{
			client:         c,
			secretsService: secretsService,
			log:            logger,
		}
	}
	return &secretsKVStoreSQL{
		sqlStore:       sqlStore,
		secretsService: secretsService,
		log:            logger,
		decryptionCache: decryptionCache{
			cache: make(map[int64]cachedDecrypted),
		},
	}
}

// SecretsKVStore is an interface for k/v store.
type SecretsKVStore interface {
	Get(ctx context.Context, orgId int64, namespace string, typ string) (string, bool, error)
	Set(ctx context.Context, orgId int64, namespace string, typ string, value string) error
	Del(ctx context.Context, orgId int64, namespace string, typ string) error
	Keys(ctx context.Context, orgId int64, namespace string, typ string) ([]Key, error)
	Rename(ctx context.Context, orgId int64, namespace string, typ string, newNamespace string) error
}

// WithType returns a kvstore wrapper with fixed orgId and type.
func With(kv SecretsKVStore, orgId int64, namespace string, typ string) *FixedKVStore {
	return &FixedKVStore{
		kvStore:   kv,
		OrgId:     orgId,
		Namespace: namespace,
		Type:      typ,
	}
}

// FixedKVStore is a SecretsKVStore wrapper with fixed orgId, namespace and type.
type FixedKVStore struct {
	kvStore   SecretsKVStore
	OrgId     int64
	Namespace string
	Type      string
}

func (kv *FixedKVStore) Get(ctx context.Context) (string, bool, error) {
	return kv.kvStore.Get(ctx, kv.OrgId, kv.Namespace, kv.Type)
}

func (kv *FixedKVStore) Set(ctx context.Context, value string) error {
	return kv.kvStore.Set(ctx, kv.OrgId, kv.Namespace, kv.Type, value)
}

func (kv *FixedKVStore) Del(ctx context.Context) error {
	return kv.kvStore.Del(ctx, kv.OrgId, kv.Namespace, kv.Type)
}

func (kv *FixedKVStore) Keys(ctx context.Context) ([]Key, error) {
	return kv.kvStore.Keys(ctx, kv.OrgId, kv.Namespace, kv.Type)
}

func (kv *FixedKVStore) Rename(ctx context.Context, newNamespace string) error {
	err := kv.kvStore.Rename(ctx, kv.OrgId, kv.Namespace, kv.Type, newNamespace)
	if err != nil {
		return err
	}
	kv.Namespace = newNamespace
	return nil
}
