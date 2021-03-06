package httpapi

import (
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/serf/serf"
)

func GetConsulClient() (*api.Client, error) {
	return api.NewClient(api.DefaultNonPooledConfig())
}

type MemberAgent struct {
	HostName string `json:"hostName"`
	Address  string `json:"address"`
	// TODO: rename to status and give more choise
	Avail bool `json:"avail"`
}

func GetConsulAgents(client *api.Client) ([]MemberAgent, error) {
	members, err := client.Agent().Members(false)
	if err != nil {
		return nil, err
	}
	result := make([]MemberAgent, 0)
	for _, m := range members {
		status := serf.MemberStatus(m.Status)
		if status == serf.StatusLeaving || status == serf.StatusLeft {
			continue
		}
		// TODO: return status not just bool value
		available := status == serf.StatusAlive
		result = append(result, MemberAgent{m.Name, m.Addr, available})
	}
	return result, nil
}

func JoinConsulAgent(client *api.Client, address string) error {
	return client.Agent().Join(address, false)
}

func ForceLeaveConsulAgent(client *api.Client, nodename string) error {
	return client.Agent().ForceLeave(nodename)
}
