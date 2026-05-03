Build Wavesmith and launch it. Steps:

1. Kill any running Wavesmith instance: `pkill -x Wavesmith` (ignore errors if not running)
2. Build the Xcode project:
   ```
   xcodebuild -project native/WavesmithNative/WavesmithNative.xcodeproj -scheme WavesmithNative build 2>&1 | tail -5
   ```
3. If build succeeds, launch:
   ```
   open ~/Library/Developer/Xcode/DerivedData/WavesmithNative-*/Build/Products/Debug/Wavesmith.app
   ```
4. Report success or failure.

Do NOT ask for confirmation — just build and launch.
