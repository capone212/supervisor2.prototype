package httpapi

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	return nil
}

func ListInterfaces(w http.ResponseWriter, r *http.Request) error {
	interfcaes, err := net.InterfaceAddrs()
	if err != nil {
		return err
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
	return nil
}

func ListAgentMembers(w http.ResponseWriter, r *http.Request) error {
	client, err := GetConsulClient()
	if err != nil {
		return err
	}
	var members []MemberAgent
	members, err = GetConsulAgents(client)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(members); err != nil {
		panic(err)
	}
	return err
}
