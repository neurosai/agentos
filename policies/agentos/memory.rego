package agentos.memory

default allow := false

allow if {
	input.action == "write"
}

allow if {
	input.action == "delete"
}

# Query/list gate — no specific record yet.
allow if {
	input.action == "search"
	"engineer" in input.subject.roles
}

allow if {
	input.action == "read"
	input.record.classification == "internal"
	input.record.namespace == "workspace:payments"
	input.subject.group_ids[_] == "group:team-payments"
}

allow if {
	input.action == "read"
	startswith(input.record.namespace, "catalog:")
}

allow if {
	input.action == "read"
	input.record.classification == "internal"
}
