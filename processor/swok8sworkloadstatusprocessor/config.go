package swok8sworkloadstatusprocessor

import (
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`
	WatchSyncPeriod     time.Duration `mapstructure:"watch_sync_period"`
}

func (c *Config) getK8sClient() (kubernetes.Interface, error) {
	return initKubeClient(c.APIConfig)
}
