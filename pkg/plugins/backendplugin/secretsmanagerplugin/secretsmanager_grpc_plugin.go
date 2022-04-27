package secretsmanagerplugin

type SecretsManagerPlugin interface {
	RemoteSecretsManagerClient
}

// type SecretsManagerGRPCPlugin struct {
// 	plugin.NetRPCUnsupportedPlugin
// }

// func (p *SecretsManagerGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
// 	return nil
// }

// func (p *SecretsManagerGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
// 	return &SecretsManagerGRPCClient{NewRemoteSecretsManagerClient(c)}, nil
// }

// type SecretsManagerGRPCClient struct {
// 	RemoteSecretsManagerClient
// }

// func (m *SecretsManagerGRPCClient) Get(ctx context.Context, req *SecretsRequest, opts ...grpc.CallOption) (*SecretsGetResponse, error) {
// 	return m.RemoteSecretsManagerClient.Get(ctx, req)
// }

// var _ RemoteSecretsManagerClient = &SecretsManagerGRPCClient{}
// var _ plugin.GRPCPlugin = &SecretsManagerGRPCPlugin{}
