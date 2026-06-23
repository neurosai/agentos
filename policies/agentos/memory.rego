package agentos.memory

default allow := false

allow if {
	input.action == "write"
}

allow if {
	input.action == "delete"
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
