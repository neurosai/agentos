package agentos.tasks_test

import rego.v1

test_task_submit_allowed if {
	agentos.tasks.allow with input as {
		"action": "task.submit",
		"subject": {"roles": ["engineer"]},
	}
}

test_task_read_allowed if {
	agentos.tasks.allow with input as {
		"action": "task.read",
	}
}
