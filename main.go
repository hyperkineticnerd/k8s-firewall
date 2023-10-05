package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hyperkineticnerd/k8s-firewall/client"
	"github.com/hyperkineticnerd/k8s-firewall/config"
	"github.com/hyperkineticnerd/k8s-firewall/controller"
	"github.com/hyperkineticnerd/k8s-firewall/provider/juniper"
	"github.com/hyperkineticnerd/k8s-firewall/source"
	"github.com/hyperkineticnerd/k8s-firewall/templates"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.NewConfig()
	Flags(cfg)
	Logging(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	go handleSigterm(cancel)
	ctrl := Controller(ctx, cfg)
	defer ctrl.Router.Session.Close()
	ctrl.RunOnce(ctx)
}

func handleSigterm(cancel func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	<-signals
	logrus.Info("Received SIGTERM. Termination...")
	cancel()
}

func Logging(cfg *config.Config) {
	ll, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Fatalf("faile to parse log level: %v", err)
	}
	logrus.SetLevel(ll)
}

func Flags(cfg *config.Config) {
	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		logrus.Fatalf("flag parsing error: %v", err)
	}
}

func Controller(ctx context.Context, cfg *config.Config) *controller.Controller {
	cg := client.NewClientGenerator(cfg)
	tmpl := templates.TemplateSetup(cfg)
	provider := juniper.JuniperSetup(cfg)
	src, err := source.BuildWithConfig(ctx, "service", cg, cfg)
	if err != nil {
		logrus.Error(err)
	}
	return &controller.Controller{
		Client:         cg,
		Source:         src,
		Router:         *provider,
		TemplateEngine: *tmpl,
	}
}
