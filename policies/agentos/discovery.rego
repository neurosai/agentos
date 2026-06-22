package agentos.discovery

default allow := false

unsafe_collectors := {
	"packet_capture",
	"host_scan",
	"credential_guess",
	"secret_read",
	"network_sniff",
}

allow if {
	input.action == "request"
	input.collector == input.collector
	not input.collector in unsafe_collectors
	input.mode == "read_only"
}

deny if {
	input.collector in unsafe_collectors
}

deny if {
	input.mode != "read_only"
}
