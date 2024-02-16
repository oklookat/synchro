# Requirements

OS: Windows 10; macOS Catalina; Linux 5.4 (?); Android 7.0; iOS 13.

# Requirements dev

1. Non-Windows: GCC
2. Windows: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download)
3. For mobile build: `go install golang.org/x/mobile/cmd/gomobile@latest`

# Native (Go)

Just copy this project to your Go project. Then use the packages [you need](../commander).

# Shared (.dll, .so) (NOT WORKING NOW)

Windows:

[build_dll.cmd](./build_dll.cmd)

Linux:

[build_so.cmd](./build_so.sh)

macOS:

???

# Android (.aar)

Requirements: Android Studio (ANDROID_HOME, SDK, NDK).

https://github.com/zchee/golang-wiki/blob/master/Mobile.md#building-and-deploying-to-android-1

Windows: [build_android.cmd](./build_android.cmd)

Linux: [build_android.sh](./build_android.sh)

Import .aar (Android Studio):

1. Create _libs_ dir in _ANDROID_PROJECT_DIR/app_
2. Copy _core.aar_ to _libs_ dir
3. Open Android Studio.
4. File -> Project Structure -> Dependencies -> app -> + -> JAR/AAR -> _libs/core.aar_.

# iOS (.framework) (not tested)

Requirements: macOS, Xcode.

https://github.com/zchee/golang-wiki/blob/master/Mobile.md#building-and-deploying-to-ios-1

[build_ios.sh](./build_ios.sh)

# Manuals (gomobile)

https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile

https://github.com/zchee/golang-wiki/blob/master/Mobile.md
