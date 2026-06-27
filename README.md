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

Bundled extensions are **not** passed via `--load-extension` (removed in Chrome 137+). During packaging, `tools/package` calls `tools/install-ext` (Node + chrome-launcher) to run CDP `Extensions.loadUnpacked`, persisting extensions under `Data/Default/Extensions/`.

| Build-time | Runtime |
|------------|---------|
| Download CRX → unpack to `Extensions/{id}/` | Extensions load from profile |
| Node CDP install into `Data/` | No Web Store or `--load-extension` needed |

Also disables sync, component update, background networking, Safe Browsing checks, and translate — suited for machines **without access to Google**.

**Note:** Visiting normal websites still uses the open internet. Only Google-specific background services are reduced. Safe Browsing is off — be cautious on unknown sites.

Packaging happens at build time only; the launcher does not download at runtime.

## Build & packaging

| Step | What |
| ---- | ---- |
| Generate icon | `go run ./tools/genicon` |
| Build launcher | `go build -ldflags "-H=windowsgui" -o Chrome.exe ./cmd/chrome` |
| Bundle | `npm ci` in `tools/install-ext`, then `go run ./tools/package -root dist/chrome-portable` |

Requires Google Chrome and Node.js on the build machine (`scripts/install-chrome.ps1` on Windows).

## Project structure

```
cmd/chrome/              Entry point → portable.Run()
internal/
  bundle/                Portable root, layout, Chrome CLI flags
  portable/              Runtime: launch browser
  install/               Build-time: copy Chrome, Node CDP extension install, warmup
  httpclient/            Shared User-Agent (genicon)
tools/
  genicon/               Windows icon resources (.syso)
  install-ext/           Node CDP extension installer (build-time only)
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
npm ci --prefix tools/install-ext
New-Item -ItemType Directory -Force -Path dist/chrome-portable | Out-Null
go build -ldflags "-H=windowsgui" -o dist/chrome-portable/Chrome.exe ./cmd/chrome
go run ./tools/package -root dist/chrome-portable
```

## Release

`.github/workflows/release.yml` — push to `main`, weekly schedule, or manual dispatch. Publishes `chrome-portable-{platform}.zip`.

## License

MIT
