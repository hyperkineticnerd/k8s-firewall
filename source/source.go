package source

import (
	"context"
	"errors"

	"github.com/hyperkineticnerd/k8s-firewall/client"
	"github.com/hyperkineticnerd/k8s-firewall/config"
	"github.com/hyperkineticnerd/k8s-firewall/provider/juniper"
)

type Source interface {
	PortForwards(ctx context.Context) ([]*juniper.PortForward, error)
}

var ErrSourceNotFound = errors.New("source not found")

func BuildWithConfig(ctx context.Context, source string, p client.ClientGenerator, cfg *config.Config) (Source, error) {
	switch source {
	case "service":
		client, err := p.KubeClient()
		if err != nil {
			return nil, err
		}
		return NewServiceSource(ctx, client, cfg.Namespace)
	}
	return nil, ErrSourceNotFound
}
