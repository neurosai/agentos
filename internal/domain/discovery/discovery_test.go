package discovery

import (
	"errors"
	"testing"
)

func TestValidateJobSafe(t *testing.T) {
	t.Parallel()

	job := Job{
		Collector:      CollectorGit,
		Mode:           ModeReadOnly,
		WriteToCatalog: true,
		WriteToMemory:  false,
	}
	if err := ValidateJob(job); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestValidateJobUnsafeCollector(t *testing.T) {
	t.Parallel()

	if !UnsafeCollector("packet_capture") {
		t.Fatal("packet_capture should be unsafe")
	}
}

func TestValidateJobMemoryWithoutCatalog(t *testing.T) {
	t.Parallel()

	job := Job{
		Collector:      CollectorKubernetes,
		Mode:           ModeReadOnly,
		WriteToCatalog: false,
		WriteToMemory:  true,
	}
	if err := ValidateJob(job); !errors.Is(err, ErrMemoryWithoutCatalog) {
		t.Fatalf("got %v", err)
	}
}
