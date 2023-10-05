package config

import (
	"fmt"
	"reflect"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
)

var Version = "unknown"

const (
	passwordMask = "*******"
)

type Config struct {
	APIServerURL         string
	KubeConfig           string
	RequestTimeout       time.Duration
	Sources              []string
	Namespace            string
	AnnotationFilter     string
	LabelFilter          string
	Interval             time.Duration
	MinEventSyncInterval time.Duration
	Once                 bool
	DryRun               bool
	UpdateEvents         bool
	LogLevel             string
	SSHUsername          string
	SSHPassphrase        string
	RouterHost           string
	TemplatePath         string
	TemplateName         string
}

var defaultConfig = &Config{
	APIServerURL:         "",
	KubeConfig:           "",
	RequestTimeout:       time.Second * 30,
	Sources:              nil,
	Namespace:            "",
	AnnotationFilter:     "",
	LabelFilter:          "",
	Interval:             time.Minute,
	MinEventSyncInterval: 5 * time.Second,
	Once:                 false,
	DryRun:               false,
	UpdateEvents:         false,
	LogLevel:             logrus.InfoLevel.String(),
	SSHUsername:          "root",
	SSHPassphrase:        "",
	RouterHost:           "",
	TemplatePath:         "",
	TemplateName:         "portforward.tmpl",
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) String() string {
	temp := *cfg
	t := reflect.TypeOf(temp)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if val, ok := f.Tag.Lookup("secure"); ok && val == "yes" {
			if f.Type.Kind() != reflect.String {
				continue
			}
			v := reflect.ValueOf(&temp).Elem().Field(i)
			if v.String() != "" {
				v.SetString(passwordMask)
			}
		}
	}
	return fmt.Sprintf("%+v", temp)
}

func allLogLevelsAsStrings() []string {
	var levels []string
	for _, level := range logrus.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

func (cfg *Config) ParseFlags(args []string) error {
	app := kingpin.New("k8s-firewall", "External Firewall configuration from within Kubernetes")
	app.Version(Version)
	app.DefaultEnvars()

	// Kubernetes Flags
	app.Flag("server", "The kubernetes API server to connect to (default: auto-detect)").Default(defaultConfig.APIServerURL).StringVar(&cfg.APIServerURL)
	app.Flag("kubeconfig", "Retrieve target cluster configuration from a kubeconfig").Default(defaultConfig.KubeConfig).StringVar(&cfg.KubeConfig)
	app.Flag("request-timeout", "Set K8s Request Timeout").Default(defaultConfig.RequestTimeout.String()).DurationVar(&cfg.RequestTimeout)

	// Juniper Flags
	app.Flag("router", "Router address to SSH to").Default(defaultConfig.RouterHost).StringVar(&cfg.RouterHost)
	app.Flag("user", "The SSH Username to access the Juniper SRX").Default(defaultConfig.SSHUsername).StringVar(&cfg.SSHUsername)
	app.Flag("pass", "The SSH Passphrase to access the Juniper SRX").Default(defaultConfig.SSHPassphrase).StringVar(&cfg.SSHPassphrase)

	// misc Flags
	app.Flag("log-level", "Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal").Default(defaultConfig.LogLevel).EnumVar(&cfg.LogLevel, allLogLevelsAsStrings()...)
	app.Flag("template-path", "Template Path").Default(defaultConfig.TemplatePath).StringVar(&cfg.TemplatePath)
	app.Flag("template-name", "Template Name").Default(defaultConfig.TemplateName).StringVar(&cfg.TemplateName)
	_, err := app.Parse(args)
	if err != nil {
		return err
	}
	return nil
}
