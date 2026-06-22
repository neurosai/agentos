package agentos.memory_test

import rego.v1

import data.agentos.memory

test_allow_payments_workspace if {
	memory.allow with input as {
		"action": "read",
		"record": {
			"classification": "internal",
			"namespace": "workspace:payments",
		},
		"subject": {"group_ids": ["group:team-payments"]},
	}
}

test_allow_catalog_namespace if {
	memory.allow with input as {
		"action": "read",
		"record": {"namespace": "catalog:payments"},
		"subject": {"group_ids": []},
	}
}
