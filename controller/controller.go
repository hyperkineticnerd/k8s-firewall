package controller

import (
	"context"
	"sync"
	"time"

	"github.com/hyperkineticnerd/k8s-firewall/client"
	"github.com/hyperkineticnerd/k8s-firewall/provider/juniper"
	"github.com/hyperkineticnerd/k8s-firewall/source"
	"github.com/hyperkineticnerd/k8s-firewall/templates"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	Client               client.ClientGenerator
	Source               source.Source
	Router               juniper.JuniperConnection
	TemplateEngine       templates.TemplateEngine
	Interval             time.Duration
	nextRunAt            time.Time
	nextRunAtMux         sync.Mutex
	MinEventSyncInterval time.Duration
}

func (c *Controller) RunOnce(ctx context.Context) error {
	portforwards, err := c.Source.PortForwards(ctx)
	if err != nil {
		return err
	}
	if len(portforwards) > 0 {
		for _, portforward := range portforwards {
			routerConfig, _ := c.TemplateEngine.TemplateRender(portforward)
			c.Router.Edit(routerConfig)
			return nil
		}
	}
	return nil
}

func (c *Controller) ScheduleRunOnce(now time.Time) {
	c.nextRunAtMux.Lock()
	defer c.nextRunAtMux.Unlock()
	if !c.nextRunAt.Before(now.Add(c.MinEventSyncInterval)) {
		c.nextRunAt = now.Add(c.MinEventSyncInterval)
	}
}

func (c *Controller) ShouldRunOnce(now time.Time) bool {
	c.nextRunAtMux.Lock()
	defer c.nextRunAtMux.Unlock()
	if now.Before(c.nextRunAt) {
		return false
	}
	c.nextRunAt = now.Add(c.Interval)
	return true
}

func (c *Controller) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if c.ShouldRunOnce(time.Now()) {
			if err := c.RunOnce(ctx); err != nil {
				logrus.Error(err)
			}
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			logrus.Info("Terminating main controller loop")
			return
		}
	}
}
