# chrome-portable

Windows-only portable Google Chrome Stable. A small Go launcher (`Chrome.exe`) starts a bundled browser and keeps all profile data inside the same folder. Copy the whole directory to move it.

## Portable layout

```
./
├── Chrome.exe      # Go launcher (double-click)
├── Chrome/         # Bundled browser (chrome.exe + DLLs)
├── Extensions/     # CopyTables + Constellation Mix (source; installed into Data/ at build time)
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

Bundled extensions are **not** passed via `--load-extension` (removed in Chrome 137+). During packaging, `tools/package` saves each CRX under `Extensions/{id}/extension.crx`, writes `Data/External Extensions/{id}.json`, then launches Chrome headless once so extensions are copied into `Data/Default/Extensions/`.

| Build-time | Runtime |
|------------|---------|
| Download CRX → `Extensions/{id}/` | Extensions load from profile |
| External Extensions JSON + headless warmup | No Web Store or `--load-extension` needed |

Also disables sync, component update, background networking, Safe Browsing checks, and translate — suited for machines **without access to Google**.

**Note:** Visiting normal websites still uses the open internet. Only Google-specific background services are reduced. Safe Browsing is off — be cautious on unknown sites.

Packaging happens at build time only; the launcher does not download at runtime.

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
  install/               Build-time: copy Chrome, external CRX install, profile warmup
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
