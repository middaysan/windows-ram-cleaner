// internal/tray/tray.go

// Description: This file contains the tray package which is responsible for creating the system tray icon and handling the tray menu items.
package tray

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"
)

// Embed the icon using go:embed
//
//go:embed assets/icon.ico
var iconData []byte

var LastCleanup time.Time

func OnReady(emptyStandbyList func() error, checkAndCleanStandbyList func()) {
	systray.SetIcon(iconData) // Use the embedded icon
	systray.SetTitle("Memory Cleaner")
	systray.SetTooltip("Right-click to clean standby list")

	mClean := systray.AddMenuItem("Clean", "Clean the standby list")
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	go func() {
		for {
			select {
			case <-mClean.ClickedCh:
				if err := emptyStandbyList(); err != nil {
					log.Printf("Error cleaning standby list: %v\n", err)
				} else {
					log.Println("Standby list cleaned successfully")
					LastCleanup = time.Now()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()

	// Periodically update the tooltip with memory information
	go func() {
		for {
			checkAndCleanStandbyList()
			time.Sleep(1 * time.Minute) // Update every minute
		}
	}()
}

func UpdateTooltip(standbySize, freeSize uint64, percent uint64) {
	tooltip := fmt.Sprintf("Standby List: %d MB, Free Memory: %d MB, Percent: %d%%", standbySize/1024/1024, freeSize/1024/1024, percent)
	systray.SetTooltip(tooltip)
}
