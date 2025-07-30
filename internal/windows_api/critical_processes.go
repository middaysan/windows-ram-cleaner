package windowsapi

import "golang.org/x/sys/windows"

// CriticalProcesses is a map of critical processes that should not be cleaned
// Use empty struct as value to save memory
var CriticalProcesses = map[string]struct{}{
	"csrss.exe":                   {},
	"wininit.exe":                 {},
	"services.exe":                {},
	"lsass.exe":                   {},
	"winlogon.exe":                {},
	"explorer.exe":                {},
	"smss.exe":                    {},
	"svchost.exe":                 {},
	"System":                      {},
	"System Idle Process":         {},
	"conhost.exe":                 {},
	"dwm.exe":                     {},
	"taskhost.exe":                {},
	"taskhostw.exe":               {},
	"spoolsv.exe":                 {},
	"msmpeng.exe":                 {},
	"audiodg.exe":                 {},
	"fontdrvhost.exe":             {},
	"sihost.exe":                  {},
	"dllhost.exe":                 {},
	"logonui.exe":                 {},
	"lsm.exe":                     {},
	"SearchIndexer.exe":           {},
	"SecurityHealthService.exe":   {},
	"ShellExperienceHost.exe":     {},
	"StartMenuExperienceHost.exe": {},
	"SystemSettings.exe":          {},
	"taskeng.exe":                 {},
	"taskhostex.exe":              {},
	"TrustedInstaller.exe":        {},
	"userinit.exe":                {},
	"WmiPrvSE.exe":                {},
	"WUDFHost.exe":                {},
}

// isCriticalProcess checks if a process is critical and should not be cleaned
func isCriticalProcess(pe windows.ProcessEntry32) bool {
	appExeFilename := windows.UTF16ToString(pe.ExeFile[:])
	_, isCritical := CriticalProcesses[appExeFilename]
	return isCritical
}
