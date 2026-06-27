# chrome-portable

Windows-only portable Google Chrome Stable. A small Go launcher (`Chrome.exe`) starts a bundled browser and keeps all profile data inside the same folder. Copy the whole directory to move it.

## Portable layout

```
./
├── Chrome.exe      # Go launcher (double-click)
├── Chrome/         # Bundled browser (chrome.exe + DLLs)
├── Extensions/     # CopyTables + Constellation Mix (manual install, see below)
│   ├── ekdpkppgmlalfkphpibadldikjimijon/
│   └── kdnbphngjjojcnnapaegdgjpgadlhbke/
├── Data/           # User profile
├── cache/          # Disk / media / GPU caches
└── crash/          # Crash dumps
```

The portable root is always the directory that contains `Chrome.exe`.

## Runtime logic

Double-clicking `Chrome.exe`:

1. Resolves the portable root from the launcher path (`bundle.Root`).
2. Verifies `Chrome/chrome.exe` exists.
3. Creates `Data/` if missing.
4. Launches `Chrome/chrome.exe` with portable flags (`--user-data-dir`, cache dirs, offline-friendly options).

Also disables sync, component update, background networking, Safe Browsing checks, and translate — suited for machines **without access to Google**.

**Note:** Visiting normal websites still uses the open internet. Only Google-specific background services are reduced. Safe Browsing is off — be cautious on unknown sites.

Packaging happens at build time only; the launcher does not download at runtime.

## Extensions (manual install)

Chrome 137+ removed `--load-extension`. Extensions are bundled under `Extensions/` at build time but **not** installed automatically. After first launch:

1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Click **Load unpacked**
4. Select each folder under `Extensions/`:
   - `Extensions/ekdpkppgmlalfkphpibadldikjimijon` — CopyTables
   - `Extensions/kdnbphngjjojcnnapaegdgjpgadlhbke` — Constellation Mix

Once loaded, extensions are stored in `Data/` and persist across launches.

## Build & packaging

| Step | What |
| ---- | ---- |
| Generate icon | `go run ./tools/genicon` |
| Build launcher | `go build -ldflags "-H=windowsgui" -o Chrome.exe ./cmd/chrome` |
| Bundle | `go run ./tools/package -root dist/chrome-portable` (trims locales, old versions) |

Requires Google Chrome on the build machine (`scripts/install-chrome.ps1` on Windows).

**Size:** Packaging removes old Chrome version folders, unused locale `.pak` files (keeps en-US / zh-CN / zh-TW), and skips pre-built profile data. First launch creates `Data/`.

## Project structure

```
cmd/chrome/              Entry point → portable.Run()
internal/
  bundle/                Portable root, layout, Chrome CLI flags
  portable/              Runtime: launch browser
  install/               Build-time: copy Chrome, trim size, bundle extensions
  httpclient/            Shared User-Agent (genicon)
tools/
  genicon/               Windows icon resources (.syso)
  package/               Assemble dist/chrome-portable/
scripts/
  install-chrome.ps1     Install latest Chrome on Windows
  build.ps1              Build Chrome.exe for amd64 / arm64
.github/workflows/
  release.yml            Build, release zip, cleanup workflow runs
```

## Local build

```powershell
go run ./tools/genicon
New-Item -ItemType Directory -Force -Path dist/chrome-portable | Out-Null
go build -ldflags "-H=windowsgui" -o dist/chrome-portable/Chrome.exe ./cmd/chrome
go run ./tools/package -root dist/chrome-portable
```

## Release

`.github/workflows/release.yml` — push to `main`, weekly schedule, or manual dispatch. Publishes `chrome-portable-{platform}.zip`.

## License

MIT
