package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"net/http"
	"os"
	"time"
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

func tryBecameLeader(client *api.Client, checkId string) (bool, string) {
	hostname, _ := os.Hostname()
	writeOptions := &api.WriteOptions{}
	ngpLeader := api.KVPair{Key: "ngpleader", Value: []byte(hostname)}
	entry := &api.SessionEntry{Checks: []string{"serfHealth", checkId}}
	session, _, sessionError := client.Session().Create(entry, writeOptions)
	if sessionError != nil {
		fmt.Println("Failed to acquire session", sessionError)
	}
	ngpLeader.Session = session
	writen, _, writeError := client.KV().Acquire(&ngpLeader, writeOptions)

	if writeError != nil {
		fmt.Println("Failed to became a leader:", writeError)
	}
	if !writen {
		client.Session().Destroy(session, nil)
	}

	return writen, session
}

func registerCheckHandler(sessionId string, port int) string {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}
	http.HandleFunc(fmt.Sprintf("/%s", sessionId), handler)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return fmt.Sprintf("http://127.0.0.1:%d/%s", port, sessionId)
}

func registerSupervisorCheck(client *api.Client, checkUrl string) string {
	fmt.Println("registring check for url", checkUrl)
	const SUPERVISOR_SERVICE_ID = "supervisor2"
	const SUPERVISOR_CHEK_ID = "supervisor2_check"
	agent := client.Agent()
	agent.CheckDeregister(SUPERVISOR_CHEK_ID)
	check := api.AgentCheckRegistration{ID: SUPERVISOR_CHEK_ID,
		Name:      SUPERVISOR_CHEK_ID,
		ServiceID: SUPERVISOR_SERVICE_ID}
	check.HTTP = checkUrl
	check.Interval = "1s"
	check.Status = "passing"
	e := agent.CheckRegister(&check)
	if e != nil {
		fmt.Println("Failed register supervisor check")
		panic(e)
	}
	return SUPERVISOR_CHEK_ID
}

func main() {
	fmt.Println("HelloWorld!")
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	//fmt.Println("Hello"client.Status().Leader())
	leader, leaderError := client.Status().Leader()
	peers, peersError := client.Status().Peers()

	fmt.Println("Leader", leader, leaderError)
	fmt.Println("Peers", peers, peersError)

	randomSequense := randSeq(64)
	checkId := registerSupervisorCheck(client, registerCheckHandler(randomSequense, 8081))
	defer client.Agent().CheckDeregister(checkId)
	isLeader, session := tryBecameLeader(client, checkId)
	fmt.Println("I am a leader: ", isLeader)

	if isLeader {
		var b []byte = make([]byte, 1)
		os.Stdin.Read(b)
	}
	_, er := client.Session().Destroy(session, nil)
	if er != nil {
		fmt.Println("Failed clean session", er)
	}
}
