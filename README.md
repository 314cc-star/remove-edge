# RemoveEdge

Free, open-source tool to uninstall Microsoft Edge (Chromium) from Windows 10 and 11, even when the Uninstall button in Settings is greyed out.

**Website:** https://YOUR_PROJECT.vercel.app

## Download

- [RemoveEdge-x64.exe](https://github.com/314cc-star/remove-edge/releases/latest/download/RemoveEdge-x64.exe) (most PCs)
- [RemoveEdge-arm64.exe](https://github.com/314cc-star/remove-edge/releases/latest/download/RemoveEdge-arm64.exe) (Snapdragon / Surface Pro X)

All releases: https://github.com/314cc-star/remove-edge/releases

## How to use

1. Download the `.exe` for your PC.
2. Double-click it. If SmartScreen appears, click **More info** → **Run anyway**.
3. Click **Yes** in the UAC (admin) prompt.
4. Type `Y` and press Enter, then wait for `[OK]`.
5. Restart your PC.

## What it does

1. Closes `msedge.exe`, `MicrosoftEdgeUpdate.exe` and `identity_helper.exe`.
2. Sets `EdgeUpdateDev\AllowUninstall = 1` so Edge's setup will uninstall outside the EEA.
3. Runs Edge's own `setup.exe --uninstall --force-uninstall` (system-level and per-user).
4. Removes leftover Edge shortcuts.
5. Sets the `DoNotUpdateToEdgeWithChromium` policy to block reinstall via Windows Update.

**Not touched:** Microsoft Edge WebView2 Runtime. Teams, Office and many other apps need it.

## Build from source

Requires Go 1.22+.

```
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o RemoveEdge-x64.exe .
GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o RemoveEdge-arm64.exe .
```

Pushing a tag like `v1.1` runs [.github/workflows/release.yml](.github/workflows/release.yml), which builds both binaries and publishes a GitHub Release.

## Get Edge back

Download the official installer from https://www.microsoft.com/edge.

## Support

If RemoveEdge saved you some time: [☕ Support on Ko-fi](https://ko-fi.com/314ccstar)

## License

MIT. Provided "as is", without warranty. Use at your own risk.
RemoveEdge is not affiliated with or endorsed by Microsoft. Microsoft Edge and Windows are trademarks of Microsoft Corporation.
