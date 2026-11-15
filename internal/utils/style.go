package utils

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

var (
	heading = lipgloss.NewStyle().
		Foreground(compat.CompleteColor{TrueColor: lipgloss.Color("#007aff"), ANSI256: lipgloss.Color("27")}).
		Bold(true).
		Padding(1)
	// Renders the primary section heading for CLI output.
	H1 = heading
	// Renders the secondary section heading for CLI output.
	H2 = heading.SetString("▌")

	// Renders status messages shown to the user.
	Msg = lipgloss.NewStyle().
		Bold(true).
		Padding(1)

	item = lipgloss.NewStyle().
		PaddingLeft(2)
	// Renders a success marker used in Dock summaries.
	CheckedItem = item.
			Foreground(compat.CompleteColor{TrueColor: lipgloss.Color("#63b946"), ANSI256: lipgloss.Color("41")}).
			SetString("✔︎")
)
