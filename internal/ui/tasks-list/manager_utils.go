package ui

import "github.com/dmitriy-rs/rollercoaster/internal/manager"

// ShouldShowManagerIndicator returns true if manager indicators should be shown.
// Manager indicators are only shown when there are multiple different managers.
func ShouldShowManagerIndicator(managerTitles []manager.Title) bool {
	if len(managerTitles) <= 1 {
		return false
	}

	// Check if all managers have the same name
	if len(managerTitles) == 0 {
		return false
	}

	firstName := managerTitles[0].Name
	for _, title := range managerTitles[1:] {
		if title.Name != firstName {
			return true // Found different manager names, show indicators
		}
	}

	return false // All managers have the same name, don't show indicators
}
