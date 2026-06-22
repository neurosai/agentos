package agentos.discovery_test

import rego.v1

import data.agentos.discovery

test_allow_git_collector if {
	discovery.allow with input as {
		"action": "request",
		"collector": "git",
		"mode": "read_only",
	}
}

test_deny_packet_capture if {
	discovery.deny with input as {
		"collector": "packet_capture",
		"mode": "read_only",
	}
}
