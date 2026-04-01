# Changelog

## v0.2.2 — 2026-03-31

### Added
- `pie inspect <id>` — view full request details from CLI
- `pie replay <id>` — replay webhooks from CLI
- `pie traces <tunnel>` — pipeline trace timelines in terminal
- `pie status --json` — JSON output for scripting
- `NO_COLOR` env var support (no-color.org standard)
- SSE/streaming pass-through (Vercel AI SDK, Ollama compatible)
- AI tool presets: `--ollama`, `--comfyui`, `--n8n`, `--tma`
- Auto-detect AI providers: Replicate, fal.ai, RunPod, Modal, OpenAI
- MCP (Model Context Protocol) JSON-RPC detection
- Auto-correlation of pipeline traces (60s window)
- Global request lookup API (`/api/requests/{id}`)
- Provider-colored badges in web dashboard

### Fixed
- `pie status` and `pie logs` auto-detect HTTPS vs HTTP server
- Duplicate auth variable in connect command

## v0.2.0 — 2026-03-31

### Added
- Pipeline auto-correlation tracker
- Server-side path matching for pipeline rules
- Auto step name from URL path
- Retro terminal error page (VT323, CRT scanlines, ASCII pie)
- Web dashboard redesign (Dracula theme, JetBrains Mono)
- Pipeline rules API (`POST /api/pipeline-rules`)
- Dashboard auth bypass for localhost

## v0.1.6 — 2026-03-31

### Added
- Friendly error page when localhost is not running
- Clean `Ctrl+C` exit message

## v0.1.5 — 2026-03-31

### Added
- Port-based subdomain cache (stable URLs between restarts)
- Interactive subdomain picker on `pie connect`
- Subdomain saved per port in `~/.pipepie/config.yaml`

### Fixed
- Reconnect preserves assigned subdomain

## v0.1.4 — 2026-03-31

### Fixed
- Server shutdown noise suppressed in `pie setup`
- Setup kills old process before starting

## v0.1.3 — 2026-03-31

### Added
- `pie setup` auto-starts systemd service
- Systemd unit includes `--auto-tls` when selected
- Interactive subdomain picker (choose own or random)
- Domain input trimmed (no trailing spaces)
- Setup kills existing pie server before running

### Fixed
- Reconnect stability — subdomain preserved in memory

## v0.1.2 — 2026-03-31

### Added
- Auto TLS option in `pie setup` (Let's Encrypt per-subdomain)
- `pie update` — self-update command
- `pie version` — shows version and checks for updates

## v0.1.1 — 2026-03-31

### Added
- Beautiful word-based subdomains (calm-fox, bold-owl)

### Fixed
- Domain trailing space in config
- Firewall section simplified in setup
- Lipgloss borders removed from setup (narrow terminal safe)

## v0.1.0 — 2026-03-31

### Initial release
- Noise NK encrypted tunneling (ChaChaPoly + BLAKE2b)
- Protobuf + zstd + yamux wire protocol
- Zero-config anonymous tunnels
- Named stable subdomains
- Web dashboard with request inspection and replay
- Pipeline tracing with trace headers
- TCP tunnels and WebSocket passthrough
- Auto TLS (Let's Encrypt, Cloudflare DNS-01, manual)
- Rate limiting (100 req/s per IP)
- `pie setup` interactive wizard (DNS, firewall, nginx, TLS, systemd)
- `pie doctor` server diagnostics
- `pie login/logout/account` multi-account management
- `pie dashboard` browser auth via Noise NK
- `pie logs --body --follow` request streaming
- `pie up` multi-tunnel from pipepie.yaml
- `pie connect --auth` password protection
- `--basic-auth` on public URLs
- huh Dracula theme for interactive CLI
- Makefile + goreleaser + cross-compile
- 36 tests (protocol, store, E2E)
