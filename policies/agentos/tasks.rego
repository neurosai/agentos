package agentos.tasks

default allow := false

allow if {
	input.action == "task.submit"
}

allow if {
	input.action == "task.read"
}

allow if {
	input.action == "task.cancel"
}

allow if {
	input.action == "task.approve"
}

deny_reason := "task action not permitted" if {
	not allow
}
