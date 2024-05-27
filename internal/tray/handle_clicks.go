// Description: This file contains functions for handling menu item clicks.

package tray

import (
	"fmt"
	"os"

	"github.com/getlantern/systray"

	winstartup "windows-ram-cleaner/internal/win_startup"
	"windows-ram-cleaner/internal/windows_api"
)

// handleMenuClicks listens for clicks on tray menu items and performs the corresponding actions.
func handleMenuClicks(TrayMenuItems *TrayMenuItems) {
	for {
		select {
		case <-TrayMenuItems.MRAMCleanForce.ClickedCh:
			handleRAMClean(true)
		case <-TrayMenuItems.MRAMCleanSafe.ClickedCh:
			handleRAMClean(false)
		case <-TrayMenuItems.MSTDClean.ClickedCh:
			handleSTDClean()
		case <-TrayMenuItems.MStartupAdd.ClickedCh:
			handleAddToStartup()
		case <-TrayMenuItems.MStartupRemove.ClickedCh:
			handleRemoveFromStartup()
		case <-TrayMenuItems.MQuit.ClickedCh:
			handleQuit()
		}
	}
}

// handleRAMClean handles RAM cleaning based on the given option.
func handleRAMClean(ignoreCritical bool) {
	options := windowsapi.CleanOptions{
		IgnoreCritical: ignoreCritical,
	}
	if err := windowsapi.CleanRAM(options); err != nil {
		windowsapi.ShowError(
			fmt.Sprintf("Can't clean RAM, err: %s", err.Error()),
			"Error cleaning RAM",
		)
	} else {
		UpdateTooltip()
	}
}

// handleSTDClean handles standby list cleaning.
func handleSTDClean() {
	if err := windowsapi.CleanStandbyList(); err != nil {
		windowsapi.ShowError(
			fmt.Sprintf("Can't clean standby list, err: %s", err.Error()),
			"Error cleaning standby list",
		)
	} else {
		UpdateTooltip()
	}
}

// handleAddToStartup handles adding the application to startup.
func handleAddToStartup() {
	if err := winstartup.CreateStartupTask(); err == nil {
		MenuItems.MStartupAdd.Disable()
		MenuItems.MStartupRemove.Enable()
	} else {
		windowsapi.ShowError(
			fmt.Sprintf("Can't create startup task, err: %s", err.Error()),
			"Error creating startup task",
		)
	}
}

// handleRemoveFromStartup handles removing the application from startup.
func handleRemoveFromStartup() {
	if err := winstartup.DeleteStartupTask(); err == nil {
		MenuItems.MStartupRemove.Disable()
		MenuItems.MStartupAdd.Enable()
	} else {
		windowsapi.ShowError(
			fmt.Sprintf("Can't delete startup task, err: %s", err.Error()),
			"Error deleting startup task",
		)
	}
}

// handleQuit handles quitting the application.
func handleQuit() {
	systray.Quit()
	os.Exit(0)
}
