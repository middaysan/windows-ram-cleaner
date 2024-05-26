
# Windows RAM Cleaner - Tray lighlightweightt application

## Overview

Windows RAM Cleaner is a lightweight tray application built with Go that helps manage and clean your system's standby memory. The application runs smoothly on Windows without any additional system requirements, such as the .NET Framework.

## Key Features

- **No Additional Requirements**: Pure Go application, no .NET Framework needed.
- **Standby Memory Cleaning**: Automatically cleans the standby memory list when it exceeds 65% of free memory.
- **System Tray Integration**: Runs quietly in the system tray with easy access.
- **Memory Usage Display**: Hover over the tray icon to view memory usage statistics.

## Requirements

- Requests administrator privileges on startup for effective memory management.
- Windows operating system

## Installation

1. Download the latest release from the [Releases page](https://github.com/middaysan/windows-ram-cleaner/releases).
2. Extract the downloaded files to your desired location.
3. Run the application by double-clicking `windows-ram-cleaner.exe`.

## Usage

1. Run the application:
    ```sh
    ./windows-ram-cleaner.exe
    ```
2. The application will appear in the system tray.
3. Right-click the tray icon to access the menu.
4. Select "Clean" to clean the standby memory list.
5. Select "Quit" to exit the application.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
