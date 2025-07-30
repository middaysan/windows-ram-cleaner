// Package tray Description: This file contains the tray package which is responsible for creating the system tray icon and handling the tray menu items.
package tray

import (
	_ "embed"
	"fmt"
	"github.com/getlantern/systray"
	winstartup "windows-ram-cleaner/internal/win_startup"
	windowsapi "windows-ram-cleaner/internal/windows_api"
)

type StartupManagement struct {
	AddToStartup      *systray.MenuItem
	RemoveFromStartup *systray.MenuItem
}

type TrayMenuItems struct {
	MSTDClean       *systray.MenuItem
	MRAMClean       *systray.MenuItem
	MRAMCleanForce  *systray.MenuItem
	MRAMCleanSafe   *systray.MenuItem
	MStartupOptions *systray.MenuItem
	MStartupAdd     *systray.MenuItem
	MStartupRemove  *systray.MenuItem
	MQuit           *systray.MenuItem
}

// Embed the icon using go:embed
//
//go:embed assets/icon.ico
var iconData []byte

// MenuItems stores the menu items for the system tray.
var MenuItems = TrayMenuItems{}

// OnReady initializes the system tray icon and menu items.
func OnReady() {
	initializeTrayIcon()
	initializeMenuItems()
	checkAndManageStartup()
	go handleMenuClicks(&MenuItems)
}

// initializeTrayIcon sets the icon and title for the system tray.
func initializeTrayIcon() {
	systray.SetIcon(iconData) // Use the embedded icon
	systray.SetTitle("Memory Cleaner")
}

// initializeMenuItems creates and sets up the menu items.
func initializeMenuItems() {
	MenuItems.MSTDClean = systray.AddMenuItem("Clean Standby List", "Clean the standby list")

	// Create a submenu for Clean RAM with Force and Safe options
	MenuItems.MRAMClean = systray.AddMenuItem("Clean RAM", "Clean the RAM")
	MenuItems.MRAMCleanSafe = MenuItems.MRAMClean.AddSubMenuItem("Basic Clean", "Basic Clean")
	MenuItems.MRAMCleanForce = MenuItems.MRAMClean.AddSubMenuItem("Deep Clean", "Thorough Clean")

	// Create a submenu for startup options
	MenuItems.MStartupOptions = systray.AddMenuItem("Startup Options", "Manage startup options")
	MenuItems.MStartupAdd = MenuItems.MStartupOptions.AddSubMenuItem("Add to Startup", "Add the application to startup")
	MenuItems.MStartupRemove = MenuItems.MStartupOptions.AddSubMenuItem("Remove from Startup", "Remove the application from startup")

	MenuItems.MQuit = systray.AddMenuItem("Quit", "Exit the application")
}

// checkAndManageStartup checks the startup task status and updates the menu items accordingly.
func checkAndManageStartup() {
	ex, err := winstartup.IsStartupTaskExists()
	if err != nil {
		MenuItems.MStartupAdd.Disable()
		MenuItems.MStartupRemove.Disable()
		windowsapi.ShowError(
			fmt.Sprintf("Can't check startup application status, err: %s", err.Error()),
			"Error checking startup task",
		)
	} else {
		if ex {
			MenuItems.MStartupAdd.Disable()
		} else {
			MenuItems.MStartupRemove.Disable()
		}
	}
}
