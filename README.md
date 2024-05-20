# Memory Cleaner Tray Application

This is a simple Windows application that runs in the system tray and allows you to clean the standby memory list. The application requests administrator privileges upon startup and provides a menu in the tray to perform memory cleaning.

## Features
- Runs in the system tray.
- Cleans the standby memory list.
- Requests administrator privileges on startup.
- Uses an embedded icon for the tray.

## Requirements
- Windows operating system.
- Go 1.16 or later.

## Installation
1. Clone the repository:
   ```sh
   git clone <repository-url>
   cd <repository-directory>
   ```

2. Install dependencies:
   ```sh
   go get -u golang.org/x/sys/windows github.com/getlantern/systray
   ```

3. Build the application:
   ```sh
   go build -ldflags="-H windowsgui -extldflags=-Wl,app.manifest"
   ```

## Usage
1. Run the application:
   ```sh
   ./memory-cleaner-tray
   ```

2. The application will appear in the system tray.
3. Right-click the tray icon to access the menu.
4. Select "Clean" to clean the standby memory list.
5. Select "Quit" to exit the application.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## License Text
```
MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
