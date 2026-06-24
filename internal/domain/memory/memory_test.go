package memory_test

import (
	"encoding/json"
	"testing"

	"github.com/neurosai/agentos/internal/domain/memory"
	"github.com/stretchr/testify/require"
)

func TestQueryJSONUnmarshal(t *testing.T) {
	var q memory.Query
	err := json.Unmarshal([]byte(`{"query":"payment","namespace":"workspace:payments","limit":10}`), &q)
	require.NoError(t, err)
	require.Equal(t, "payment", q.QueryText)
	require.Equal(t, "workspace:payments", q.Namespace)
	require.Equal(t, 10, q.Limit)
}
