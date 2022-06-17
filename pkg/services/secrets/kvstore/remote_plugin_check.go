package kvstore

import (
	"context"

	"github.com/grafana/grafana/pkg/plugins/backendplugin/secretsmanagerplugin"
)

type UseRemoteSecretsPluginCheck interface {
	ShouldUseRemoteSecretsPlugin() bool
	GetPlugin() (secretsmanagerplugin.SecretsManagerPlugin, error)
	StartPlugin(context.Context) error
}

type OSSRemoteSecretsPluginCheck struct {
	UseRemoteSecretsPluginCheck
}

func ProvideRemotePluginCheck() *OSSRemoteSecretsPluginCheck {
	return &OSSRemoteSecretsPluginCheck{}
}

func (c *OSSRemoteSecretsPluginCheck) ShouldUseRemoteSecretsPlugin() bool {
	return false
}

func (c *OSSRemoteSecretsPluginCheck) GetPlugin() (secretsmanagerplugin.SecretsManagerPlugin, error) {
	return nil, nil
}

func (c *OSSRemoteSecretsPluginCheck) StartPlugin(ctx context.Context) error {
	return nil
}
