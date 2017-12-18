package registration

import consul "github.com/hashicorp/consul/api"

func NewConsulClient(addr string) (Client, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &client{consul: c}, nil
}

type Client interface {
	Service(string, string) ([]string, error)
	Register(string, int) error
	DeRegister(string) error
}

type client struct{
	consul *consul.Client
}

func(c *client) Register(name string, port int) error {
	return nil
}

func(c *client) DeRegister(id string) error {
	return nil
}

func(c *client) Service(service, tag string) ([] *consul.ServiceEntry, *consul.QueryMeta, error) {

}