package tests

import (
	"os"
	"testing"

	"backend/storage/mounts"
)

func TestMountIDs_PersistenciaLetraYCorrelativo(t *testing.T) {
	t.Parallel()

	// Limpia estado previo para una prueba determinística (opcional).
	_ = os.Remove("/tmp/mia_mount_state.json")

	st := mounts.NewState()

	id1, err := st.NextID("84", "sig-123")
	if err != nil {
		t.Fatalf("NextID err: %v", err)
	}
	id2, err := st.NextID("84", "sig-123")
	if err != nil {
		t.Fatalf("NextID err: %v", err)
	}

	if id1 == id2 {
		t.Fatalf("IDs no deben repetirse: %s vs %s", id1, id2)
	}

	// Nueva instancia debe recordar letra y correlativo
	st2 := mounts.NewState()
	id3, err := st2.NextID("84", "sig-123")
	if err != nil {
		t.Fatalf("NextID err: %v", err)
	}

	if id3 == id2 {
		t.Fatalf("debería incrementar correlativo tras recarga: id2=%s id3=%s", id2, id3)
	}
}
