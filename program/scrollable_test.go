package program

import "testing"

func TestScrollableList_SetLength(t *testing.T) {
	s := &scrollableList{}
	s.SetLength(10)
	if s.Len() != 10 {
		t.Errorf("Len() = %d, want 10", s.Len())
	}
	s.SetLength(-1)
	if s.Len() != 0 {
		t.Errorf("SetLength(-1) should clamp to 0, got Len() = %d", s.Len())
	}
}

func TestScrollableList_moveUp_moveDown(t *testing.T) {
	s := &scrollableList{}
	s.SetLength(5)
	s.moveDown(3)
	if s.SelectedRow() != 1 {
		t.Errorf("after moveDown SelectedRow() = %d, want 1", s.SelectedRow())
	}
	s.moveUp()
	if s.SelectedRow() != 0 {
		t.Errorf("after moveUp SelectedRow() = %d, want 0", s.SelectedRow())
	}
	// moveUp at top is no-op
	s.moveUp()
	if s.SelectedRow() != 0 {
		t.Errorf("moveUp at top should stay 0, got %d", s.SelectedRow())
	}
}

func TestScrollableList_moveToTop_moveToBottom(t *testing.T) {
	s := &scrollableList{}
	s.SetLength(10)
	s.moveToBottom(3)
	if s.SelectedRow() != 9 {
		t.Errorf("SelectedRow() = %d, want 9", s.SelectedRow())
	}
	s.moveToTop()
	if s.SelectedRow() != 0 || s.ScrollOffset() != 0 {
		t.Errorf("after moveToTop: selected=%d scroll=%d", s.SelectedRow(), s.ScrollOffset())
	}
}

func TestScrollableList_adjustScrollOffset(t *testing.T) {
	s := &scrollableList{selectedRow: 5, scrollOffset: 0, length: 10}
	s.adjustScrollOffset(3)
	if s.ScrollOffset() != 3 {
		t.Errorf("ScrollOffset() = %d, want 3", s.ScrollOffset())
	}
}

func TestScrollableList_reset(t *testing.T) {
	s := &scrollableList{selectedRow: 4, scrollOffset: 2, length: 10}
	s.reset()
	if s.SelectedRow() != 0 || s.ScrollOffset() != 0 {
		t.Errorf("after reset: selected=%d scroll=%d", s.SelectedRow(), s.ScrollOffset())
	}
}

func TestScrollableList_emptyListNoPanic(t *testing.T) {
	s := &scrollableList{}
	s.SetLength(0)
	s.moveDown(5)
	s.moveToBottom(5)
	if s.SelectedRow() != 0 {
		t.Errorf("empty list should keep selectedRow 0, got %d", s.SelectedRow())
	}
}

func TestScrollableList_moveDownAtBottomNoOp(t *testing.T) {
	s := &scrollableList{}
	s.SetLength(5)
	s.moveToBottom(3)
	s.moveDown(3)
	if s.SelectedRow() != 4 {
		t.Errorf("moveDown at bottom should stay at 4, got %d", s.SelectedRow())
	}
}
