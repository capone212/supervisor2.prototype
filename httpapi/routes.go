package httpapi

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"Interfaces",
		"GET",
		"/ips",
		ListInterfaces,
	},
	Route{
		"Cluster Members",
		"GET",
		"/config/cluster",
		ListAgentMembers,
	},
}
