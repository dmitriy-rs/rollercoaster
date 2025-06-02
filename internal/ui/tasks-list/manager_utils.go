package ui

import "github.com/dmitriy-rs/rollercoaster/internal/manager"

func ShouldShowManagerIndicator(managerTasks []manager.ManagerTask) bool {
	var firstManagerName string
	showManagerIndicator := false

	for i, mgr := range managerTasks {
		if !showManagerIndicator {
			managerName := (*mgr.Manager).GetTitle().Name
			if i == 0 {
				firstManagerName = managerName
			} else if managerName != firstManagerName {
				showManagerIndicator = true
				break
			}
		}
	}
	return showManagerIndicator
}
