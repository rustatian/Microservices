package svcdiscovery

import (
	"fmt"

	stdconsul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

// Registrar registers service instance liveness information to Consul.
type Registrar struct {
	client       Client
	registration *stdconsul.AgentServiceRegistration
	logger       *logrus.Logger
}

// NewRegistrar returns a Consul Registrar acting on the provided catalog
// registration.
func NewRegistrar(client Client, r *stdconsul.AgentServiceRegistration, logger *logrus.Logger) *Registrar {

	logger.WithFields(logrus.Fields{
		"service": r.Name,
		"tags":    fmt.Sprint(r.Tags),
		"address": r.Address,
	}).Info("NewRegistrar")

	return &Registrar{
		client:       client,
		registration: r,
		logger:       logger,
	}
}

// Register implements sd.Registrar interface.
func (p *Registrar) Register() {
	if err := p.client.Register(p.registration); err != nil {
		p.logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Register error")
	} else {
		p.logger.WithFields(logrus.Fields{
			"action: ": "register",
		}).Info("Registered")
	}
}

// Deregister implements sd.Registrar interface.
func (p *Registrar) Deregister() {
	if err := p.client.Deregister(p.registration); err != nil {
		p.logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Deregister error")
	} else {
		p.logger.WithFields(logrus.Fields{
			"action: ": "deregister",
		}).Info("Deregistered")
	}
}
