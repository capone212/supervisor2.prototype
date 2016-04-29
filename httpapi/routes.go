package httpapi

import (
	"net/http"
)

type HttpHandlerFunc func(http.ResponseWriter, *http.Request) error

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc HttpHandlerFunc
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
	Route{
		"Join Cluster Member",
		"PUT",
		"/config/cluster",
		JoinAgentMember,
	},
	Route{
		"Force Leave Cluster Member",
		"DELETE",
		"/config/cluster",
		ForceLeaveAgentMember,
	},
}
