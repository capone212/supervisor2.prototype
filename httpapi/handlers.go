package httpapi

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}

func ListInterfaces(w http.ResponseWriter, r *http.Request) {
	interfcaes, err := net.InterfaceAddrs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error, %s", err)
	}
	// Get ipv4 address list
	var ipv4List = make([]string, 0)
	for _, address := range interfcaes {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipv4List = append(ipv4List, ipnet.IP.String())
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(ipv4List); err != nil {
		panic(err)
	}
}

type MemberAgent struct {
	HostName string `json:"hostName"`
	Address  string `json:"address"`
	// TODO: rename to status and give more choise
	Avail bool `json:"avail"`
}

func ListAgentMembers(w http.ResponseWriter, r *http.Request) {
	members := []MemberAgent{
		MemberAgent{"capone-hp", "192.168.1.2", true},
		MemberAgent{"dnk", "192.168.22.11", false},
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(members); err != nil {
		panic(err)
	}
}
