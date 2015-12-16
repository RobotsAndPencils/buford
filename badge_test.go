package buford_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford"
)

func TestDefaultBadge(t *testing.T) {
	b := buford.Badge{}
	if _, ok := b.Number(); ok {
		t.Errorf("Expected badge number to be omitted.")
	}
}

func TestPreserveBadge(t *testing.T) {
	b := buford.PreserveBadge
	if _, ok := b.Number(); ok {
		t.Errorf("Expected badge number to be omitted.")
	}
}

func TestClearBadge(t *testing.T) {
	b := buford.ClearBadge
	n, ok := b.Number()
	if !ok {
		t.Errorf("Expected badge to be set for removal.")
	}
	if n != 0 {
		t.Errorf("Expected badge number to be 0, got %d.", n)
	}
}

func TestNewBadge(t *testing.T) {
	b := buford.NewBadge(4)
	n, ok := b.Number()
	if !ok {
		t.Errorf("Expected badge to be set to change.")
	}
	if n != 4 {
		t.Errorf("Expected badge number to be 4, got %d.", n)
	}
}
