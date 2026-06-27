# chrome-portable

Windows-only portable Google Chrome Stable. A small Go launcher (`Chrome.exe`) starts a bundled browser and keeps all profile data inside the same folder. Copy the whole directory to move it.

## Portable layout

```
./
├── Chrome.exe      # Go launcher (double-click)
├── Chrome/         # Bundled browser (chrome.exe + DLLs)
├── Extensions/     # CopyTables + Constellation Mix (loaded via --load-extension)
│   ├── ekdpkppgmlalfkphpibadldikjimijon/
│   └── kdnbphngjjojcnnapaegdgjpgadlhbke/
└── Data/           # User profile (--user-data-dir, resolved at launch)
```

The portable root is always the directory that contains `Chrome.exe`.

## Runtime logic

Double-clicking `Chrome.exe`:

1. Resolves the portable root from the launcher path (`bundle.Root`).
2. Verifies `Chrome/chrome.exe` exists.
3. Creates `Data/` if missing.
4. Launches `Chrome/chrome.exe` with:

| Flag | Purpose |
|------|---------|
| `--user-data-dir=<Data>` | Profile directory (absolute path, resolved at launch) |
| `--no-default-browser-check` | Skip default-browser check |
| `--disable-logging` | Disable logging |
| `--disable-breakpad` | Disable crash reporting |
| `--load-extension=Extensions/{id}` | Load CopyTables and Constellation Mix (one flag per extension) |

Extensions are downloaded at package time into `Extensions/` (outside `Data/`). Each launch passes `--load-extension` with absolute paths.

Packaging happens at build time only; the launcher does not download or copy Chrome at runtime.

## Build & packaging

| Step | What |
| ---- | ---- |
| Generate icon | `go run ./tools/genicon` |
| Build launcher | `go build -ldflags "-H=windowsgui" -o Chrome.exe ./cmd/chrome` |
| Bundle | `go run ./tools/package -root dist/chrome-portable` (Chrome + extensions + warmup) |

Requires Google Chrome installed on the build machine (`scripts/install-chrome.ps1` on Windows).

## Project structure

```
cmd/chrome/              Entry point → portable.Run()
internal/
  bundle/                Portable root, layout, Chrome CLI flags
  portable/              Runtime: launch browser
  install/               Build-time: copy Chrome, profile warmup
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
