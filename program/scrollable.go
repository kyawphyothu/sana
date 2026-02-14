package program

// scrollableList holds row selection and scroll state for a list view.
// Set the list length with SetLength before calling moveDown or moveToBottom
// so bounds are correct. moveUp and moveToTop only use selectedRow/scrollOffset.
type scrollableList struct {
	selectedRow  int
	scrollOffset int
	items        []interface{}
}

// SetLength sets the number of items in the list (e.g. before navigation).
// It does not store the actual items; the slice is used only for length.
func (s *scrollableList) SetLength(n int) {
	if n < 0 {
		n = 0
	}
	if cap(s.items) >= n {
		s.items = s.items[:n]
	} else {
		s.items = make([]interface{}, n)
	}
}

// Len returns the number of items in the list.
func (s *scrollableList) Len() int {
	return len(s.items)
}

// SelectedRow returns the currently selected row index.
func (s *scrollableList) SelectedRow() int {
	return s.selectedRow
}

// ScrollOffset returns the current scroll offset (first visible row index).
func (s *scrollableList) ScrollOffset() int {
	return s.scrollOffset
}

// moveUp moves selection up one row and scrolls up if needed.
func (s *scrollableList) moveUp() {
	if s.selectedRow > 0 {
		s.selectedRow--
		if s.selectedRow < s.scrollOffset {
			s.scrollOffset = s.selectedRow
		}
	}
}

// moveDown moves selection down one row and adjusts scroll so the row stays visible.
// maxVisible is the number of rows visible in the viewport.
func (s *scrollableList) moveDown(maxVisible int) {
	n := s.Len()
	if n == 0 {
		return
	}
	maxRow := n - 1
	if s.selectedRow < maxRow {
		s.selectedRow++
		s.adjustScrollOffset(maxVisible)
	}
}

// moveToTop moves selection to the first row and resets scroll.
func (s *scrollableList) moveToTop() {
	s.selectedRow = 0
	s.scrollOffset = 0
}

// moveToBottom moves selection to the last row and adjusts scroll.
// maxVisible is the number of rows visible in the viewport.
func (s *scrollableList) moveToBottom(maxVisible int) {
	n := s.Len()
	if n == 0 {
		return
	}
	s.selectedRow = n - 1
	s.adjustScrollOffset(maxVisible)
}

// reset sets selection and scroll to zero (e.g. when data is reloaded).
func (s *scrollableList) reset() {
	s.selectedRow = 0
	s.scrollOffset = 0
}

// adjustScrollOffset keeps the selected row visible in the viewport.
// It scrolls down if the selection is below the visible area and clamps
// scrollOffset so we don't scroll past the end of the list.
func (s *scrollableList) adjustScrollOffset(maxVisible int) {
	if maxVisible <= 0 {
		return
	}
	n := s.Len()
	// Scroll down if selection moved below visible area
	if s.selectedRow >= s.scrollOffset+maxVisible {
		s.scrollOffset = s.selectedRow - maxVisible + 1
	}
	// Clamp scroll offset so we don't scroll past the end
	maxScrollOffset := n - maxVisible
	if maxScrollOffset < 0 {
		maxScrollOffset = 0
	}
	if s.scrollOffset > maxScrollOffset {
		s.scrollOffset = maxScrollOffset
	}
}
