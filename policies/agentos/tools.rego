package agentos.tools

default allow := false
default require_approval := false

allow if {
	input.action == "invoke"
	input.resource.type == "tool"
	input.resource.risk == "low"
}

allow if {
	input.action == "invoke"
	input.resource.type == "tool"
	input.resource.risk == "medium"
}

allow if {
	input.action == "invoke"
	input.resource.type == "tool"
	input.resource.id == "mcp:gitlab.read_file"
	input.context.classification == "internal"
	"engineer" in input.subject.roles
}

require_approval if {
	input.action == "invoke"
	input.resource.type == "tool"
	input.resource.risk == "high"
}

require_approval if {
	input.resource.id == "builtin:shell.exec"
}

deny_reason := "untrusted context cannot invoke high-risk tools" if {
	input.context.sourceTrust == "untrusted"
	input.resource.risk == "high"
}
