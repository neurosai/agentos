package agentos.tools_test

import rego.v1

import data.agentos.tools

test_allow_gitlab_read_file if {
	tools.allow with input as {
		"action": "invoke",
		"resource": {"type": "tool", "id": "mcp:gitlab.read_file"},
		"context": {"classification": "internal"},
		"subject": {"roles": ["engineer"]},
	}
}

test_require_approval_shell if {
	tools.require_approval with input as {
		"resource": {"id": "builtin:shell.exec"},
	}
}

test_deny_untrusted_high_risk if {
	tools.deny_reason == "untrusted context cannot invoke high-risk tools" with input as {
		"context": {"sourceTrust": "untrusted"},
		"resource": {"risk": "high"},
	}
}
