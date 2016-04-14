package supervisor2

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"time"
)

func main() {
	fmt.Println("HelloWorld!")
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(fmt.Errorf("Failed to create api client %s", err))
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
			fmt.Println("Error recived %s", err)
		case ch := <-keyboard:
			if ch == 'q' || ch == 'Q' {
				fmt.Println("Exit requested")
				lsession.Cancel()
				return
			}
		}
	}
}
