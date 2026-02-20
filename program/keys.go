package program

import tea "charm.land/bubbletea/v2"

func (m model) help() (tea.Model, tea.Cmd) {
	m.ui.previousSelected = m.ui.selected
	m.ui.selected = helpOverlay
	m.ui.overlay = overlayHelp
	return m, nil
}
