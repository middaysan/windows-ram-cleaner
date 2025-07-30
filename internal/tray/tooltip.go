// Description: This file contains functions for updating the tooltip text of the system tray icon.

package tray

import (
	"fmt"
	"github.com/getlantern/systray"
	"windows-ram-cleaner/internal/windows_api"
)

const (
	PercentThreshold = 65
)

// UpdateTooltip updates the tooltip text of the system tray icon.
// It retrieves the standby list and free RAM size using the windowsapi package,
// and formats the tooltip string with the obtained values and the last cleanup timestamps.
// The formatted tooltip string is then set as the tooltip for the system tray icon.
func UpdateTooltip() {
	standbyList, freeRAM, err := windowsapi.GetStanbyListAndFreeRAMSize()
	if err != nil {
		windowsapi.ShowError(
			fmt.Sprintf("Can't get standby list and free RAM size, err: %s", err.Error()),
			"Error getting standby list and free RAM size",
		)
	}

	tooltipStr := fmt.Sprintf(
		"FreeRAM      : %dMB\nStandby List : %d MB",
		freeRAM/(1024*1024),
		standbyList/(1024*1024),
	)

	systray.SetTooltip(tooltipStr)
}
