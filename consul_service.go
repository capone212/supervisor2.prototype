package main

import (
	"fmt"
)

type ConsulService struct {
	initialized bool
}

func MakeConsulService() *ConsulService {
	return &ConsulService{}
}

func (this *ConsulService) Bootstrap() error {
	if this.initialized {
		return fmt.Errorf("Can't bootstrap initialized service.")
	}

}

type MainConfig struct {
	BindAddress string `json:"bind_addr"`
	Server      bool   `json:"server"`
	//TODO reconnect_timeout
}

func WriteConfig(config MainConfig) error {
}
