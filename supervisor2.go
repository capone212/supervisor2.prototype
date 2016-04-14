package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	// DefaultLockWaitTime is how long we block for at a time to check if lock
	// acquisition is possible. This affects the minimum time it takes to cancel
	// a Lock acquisition.
	DefaultLockWaitTime = 15 * time.Second

	// DefaultLockRetryTime is how long we wait after a failed lock acquisition
	// before attempting to do the lock again. This is so that once a lock-delay
	// is in effect, we do not hot loop retrying the acquisition.
	DefaultLockRetryTime = 5 * time.Second

	// DefaultMonitorRetryTime is how long we wait after a failed monitor check
	// of a lock (500 response code). This allows the monitor to ride out brief
	// periods of unavailability, subject to the MonitorRetries setting in the
	// lock options which is by default set to 0, disabling this feature. This
	// affects locks and semaphores.
	DefaultMonitorRetryTime = 2 * time.Second
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type LeaderSession struct {
	// Provides ability to monitor our leadership status.
	// Client code must listen for events and change its role
	// as a leader or follower depending on recieved event.
	// The true value means we became leader, false means follower.
	// Any combination of events is possible and should be handled correctly.
	// for example true->true->false->false->true
	// Client should not stop listen for events when it became leader,
	// and has to be ready to step down when false event occures in leader state.
	statusChannel chan bool
	errorChannel  chan error

	// Leadership session. Implementation details dont use
	session string
	// Session check id. Implementation details dont use
	checkId     string
	stopChannel chan bool
	keyName     string
}

func clean(client *api.Client, leaderSession *LeaderSession) {
	fmt.Println("Cleaning channel")
	_, er := client.Session().Destroy(leaderSession.session, nil)
	if er != nil {
		fmt.Println(fmt.Errorf("Failed to clean session %s", er))
	}
	er = client.Agent().CheckDeregister(leaderSession.checkId)
	if er != nil {
		fmt.Println(fmt.Errorf("Failed to clean check %s", er))
	}
	fmt.Println("Gracefully cleaned LeaderSession")
}

func MakeLeaderElectionSession(client *api.Client) (*LeaderSession, error) {
	result := LeaderSession{keyName: "ngpleader"}
	randomSequense := randSeq(64)
	// TODO: let OS choose port here, instead of using hardcoded port number
	var er error
	result.checkId, er = registerSupervisorCheck(client, registerCheckHandler(randomSequense, 8081))
	if er != nil {
		return nil, fmt.Errorf("Failed to register supervisor check, %s", er)
	}
	result.session, er = createSession(client, result.checkId)
	if er != nil {
		return nil, fmt.Errorf("Failed to create the session, %s", er)
	}
	result.statusChannel = make(chan bool, 1)
	// Need no buffer here becouse we need strong sync to guaranty that
	// goroutine recieved signal and will not access shared data anymore
	result.stopChannel = make(chan bool)
	result.errorChannel = make(chan error)

	go leaderElectionProc(client, &result)
	return &result, nil
}

func leaderElectionProc(client *api.Client, lsession *LeaderSession) {
	defer clean(client, lsession)
	kv := client.KV()
	qOpts := &api.QueryOptions{
		WaitTime: DefaultLockWaitTime,
	}
	for {
		// Check we should stop
		select {
		case <-lsession.stopChannel:
			return
		default:
		}

		fmt.Println("Checking key %s with waitIndex %s", lsession.keyName, qOpts.WaitIndex)

		// Look for an existing lock, blocking until not taken
		pair, meta, err := kv.Get(lsession.keyName, qOpts)
		if err != nil {
			lsession.errorChannel <- fmt.Errorf("Failed to get KeyValue %s", err)
			return
		}
		// Is there is a leader?
		if pair == nil || pair.Session == "" {
			fmt.Println("No leader. Attemt to became")

			// No leader state. It seems that we can try acquire the lock
			pair, meta, err = tryBecameLeader(kv, lsession.session, lsession.keyName)
			if err != nil {
				lsession.errorChannel <- fmt.Errorf("tryBecameLeader returned error %s", err)
				return
			}
		}
		// Is there still no leader?
		if pair == nil || pair.Session == "" {
			fmt.Println("Sleeping lock delay")

			// It seems that we have lock-delay or temporary error
			// Wait and retry
			time.Sleep(DefaultLockRetryTime)
			continue
		}

		// At this point there is a leader. And we know its id
		iamLeader := false
		if pair != nil && pair.Session == lsession.session {
			iamLeader = true
		}
		if pair != nil && pair.Session != "" {
			qOpts.WaitIndex = meta.LastIndex
		}
		// TODO: check oldvalue == newValue and report
		lsession.statusChannel <- iamLeader
	}
}

func tryBecameLeader(kv *api.KV, session, keyName string) (*api.KVPair, *api.QueryMeta, error) {
	hostname, _ := os.Hostname()
	ngpLeader := api.KVPair{Key: keyName,
		Session: session,
		Value:   []byte(hostname)}
	_, _, writeError := kv.Acquire(&ngpLeader, &api.WriteOptions{})
	if writeError != nil {
		return nil, nil, writeError
	}

	return kv.Get(keyName, &api.QueryOptions{})
}

func createSession(client *api.Client, checkId string) (string, error) {
	entry := &api.SessionEntry{Checks: []string{"serfHealth", checkId}}
	session, _, sessionError := client.Session().Create(entry, &api.WriteOptions{})
	return session, sessionError
}

func registerCheckHandler(sessionId string, port int) string {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}
	http.HandleFunc(fmt.Sprintf("/%s", sessionId), handler)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return fmt.Sprintf("http://127.0.0.1:%d/%s", port, sessionId)
}

func registerSupervisorCheck(client *api.Client, checkUrl string) (string, error) {
	fmt.Println("registring check for url", checkUrl)
	const SUPERVISOR_SERVICE_ID = "supervisor2"
	const SUPERVISOR_CHEK_ID = "supervisor2_check"
	agent := client.Agent()
	err := agent.CheckDeregister(SUPERVISOR_CHEK_ID)
	if err != nil {
		fmt.Println(fmt.Errorf("Failed deregister supervisor check %s", err))
	}
	check := api.AgentCheckRegistration{ID: SUPERVISOR_CHEK_ID,
		Name:      SUPERVISOR_CHEK_ID,
		ServiceID: SUPERVISOR_SERVICE_ID}
	check.HTTP = checkUrl
	check.Interval = "1s"
	check.Status = "passing"
	err = agent.CheckRegister(&check)
	if err != nil {
		return "", fmt.Errorf("Failed register supervisor check %s", err)
	}
	return SUPERVISOR_CHEK_ID, nil
}

func charChannel() chan byte {
	result := make(chan byte)
	f := func() {
		var b []byte = make([]byte, 1)
		for {
			os.Stdin.Read(b)
			result <- b[0]
		}
	}
	go f()
	return result
}

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
				return
			}
		}
	}
}
