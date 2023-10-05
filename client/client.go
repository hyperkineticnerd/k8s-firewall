package client

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hyperkineticnerd/k8s-firewall/config"
	"github.com/linki/instrumented_http"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type SingletonClientGenerator struct {
	KubeConfig     string
	APIServerURL   string
	RequestTimeout time.Duration
	kubeClient     kubernetes.Interface
	kubeOnce       sync.Once
}

type ClientGenerator interface {
	KubeClient() (kubernetes.Interface, error)
	// OpenShiftClient() (openshift.Interface, error)
}

func NewClientGenerator(cfg *config.Config) *SingletonClientGenerator {
	return &SingletonClientGenerator{
		KubeConfig:   "",
		APIServerURL: "",
		RequestTimeout: func() time.Duration {
			if cfg.UpdateEvents {
				return 0
			}
			return cfg.RequestTimeout
		}(),
	}
}

func (p *SingletonClientGenerator) KubeClient() (kubernetes.Interface, error) {
	var err error
	p.kubeOnce.Do(func() {
		p.kubeClient, err = NewKubeClient(p.KubeConfig, p.APIServerURL, p.RequestTimeout)
	})
	return p.kubeClient, err
}

func instrumentedRESTConfig(kubeConfig, apiServerURL string, requestTimeout time.Duration) (*rest.Config, error) {
	config, err := GetRestConfig(kubeConfig, apiServerURL)
	if err != nil {
		return nil, err
	}
	config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return instrumented_http.NewTransport(rt, &instrumented_http.Callbacks{
			PathProcessor: func(path string) string {
				parts := strings.Split(path, "/")
				return parts[len(parts)-1]
			},
		})
	}
	config.Timeout = requestTimeout
	return config, nil
}

func GetRestConfig(kubeConfig, apiServerURL string) (*rest.Config, error) {
	if kubeConfig == "" {
		if _, err := os.Stat(clientcmd.RecommendedHomeFile); err == nil {
			kubeConfig = clientcmd.RecommendedHomeFile
		}
	}
	logrus.Debugf("apiServerURL: %s", apiServerURL)
	logrus.Debugf("kubeConfig: %s", kubeConfig)
	var (
		config *rest.Config
		err    error
	)
	if kubeConfig == "" {
		logrus.Infof("Using inCluster-config bbased on serviceaccount-token")
		config, err = rest.InClusterConfig()
	} else {
		logrus.Infof("Using kubeConfig")
		config, err = clientcmd.BuildConfigFromFlags(apiServerURL, kubeConfig)
	}
	if err != nil {
		return nil, err
	}
	return config, err
}

func NewKubeClient(kubeConfig, apiServerURL string, requestTimeout time.Duration) (*kubernetes.Clientset, error) {
	logrus.Infof("Instantiating new Kubernetes client")
	config, err := instrumentedRESTConfig(kubeConfig, apiServerURL, requestTimeout)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Created Kubernetes client %s", config.Host)
	return client, nil
}
