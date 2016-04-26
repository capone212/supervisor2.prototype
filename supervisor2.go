package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/serf/serf"
	"time"
)

func listMembers(client *api.Client) {
	members, err := client.Agent().Members(false)
	if err != nil {
		panic(err)
	}
	for _, m := range members {
		fmt.Printf("Agent name:%s address:%s status:%s \n", m.Name, m.Addr, serf.MemberStatus(m.Status))
	}
}

func main() {
	fmt.Println("HelloWorld!")
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(fmt.Errorf("Failed to create api client %s", err))
	}
	for {
		listMembers(client)
		fmt.Println("--------------")
		time.Sleep(time.Second * 3)
	}

	keyboard := charChannel()
	var lsession *LeaderSession = nil
	for {
		if err != nil || lsession == nil {
			fmt.Println("Reinit connection")
			lsession, err = MakeLeaderElectionSession(client)
			if err != nil {
				time.Sleep(DefaultMonitorRetryTime)
				continue
			}
		}
		select {
		case isLeader := <-lsession.statusChannel:
			fmt.Println("I am leader", isLeader)
		case err = <-lsession.errorChannel:
			fmt.Println(fmt.Sprintf("Error recived %s", err))
		case ch := <-keyboard:
			if ch == 'q' || ch == 'Q' {
				fmt.Println("Exit requested")
				lsession.Cancel()
				return
			}
		}
	}
}
