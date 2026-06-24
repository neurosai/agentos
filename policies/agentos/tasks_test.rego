package agentos.tasks_test

import rego.v1

import data.agentos.tasks

test_task_submit_allowed if {
	tasks.allow with input as {
		"action": "task.submit",
		"subject": {"roles": ["engineer"]},
	}
}

test_task_read_allowed if {
	tasks.allow with input as {
		"action": "task.read",
	}
}
