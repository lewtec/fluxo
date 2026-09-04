# Fluxo

BitTorrent client with a server-rendered web UI.

## Features

- **BitTorrent Client**: Built on Rain
- **templ UI**: HTML pages rendered in Go
- **Live updates**: Server-Sent Events for torrent list, detail, and speeds
- **Configuration**: CLI flags, environment variables, or a config file

## Architecture

### Backend (Go)

- **Rain**: BitTorrent protocol implementation
- **Cobra + Viper**: CLI and configuration
- **templ**: HTML templates compiled into the binary
- **Event Bus**: Internal events drive the SSE stream

### Frontend

- **templ** pages and fragments
- **daisyUI 5** on Tailwind 4 (compiled CSS, embedded)
- Small `EventSource` script swaps live fragments

## Building

### Prerequisites

- Go 1.24+
- Node.js 18+ (only to rebuild CSS)

### Build CSS (when templates change)

```bash
cd web
npm install
npm run build
cd ..
```

### Generate templates (when `.templ` files change)

```bash
go generate ./internal/webui
```

### Build binary

The backend embeds `web/static`:

```bash
go build -o fluxo ./cmd/fluxo
```

## Running

```bash
./fluxo
```

Then open `http://127.0.0.1:8080`.

`--dev-mode` turns off static-asset caching. `--dev-proxy` is ignored.

## Configuration

### CLI Flags

```bash
./fluxo --help
```

Key flags:

- `--api-port`: HTTP port (default: 8080)
- `--api-host`: HTTP host (default: 127.0.0.1)
- `--data-dir`: Downloads directory (default: ~/.fluxo/downloads)
- `--database`: Session database path (default: ~/.fluxo/session.db)
- `--debug`: Enable debug logging

### Environment Variables

All flags can be set via environment variables with a `FLUXO_` prefix:

```bash
export FLUXO_API_PORT=9090
export FLUXO_DATA_DIR=/mnt/torrents
./fluxo
```

### Config File

Create `fluxo.yaml` in the current directory, `~/.fluxo/`, or `/etc/fluxo/`:

```yaml
api-port: 8080
api-host: 127.0.0.1
data-dir: /mnt/torrents
database: /var/lib/fluxo/session.db
```

## HTTP

| Method | Path | Role |
| --- | --- | --- |
| GET | `/` | Torrent list |
| GET | `/add` | Add torrent form |
| POST | `/torrents` | Add magnet or `.torrent` file |
| GET | `/torrents/{id}` | Torrent detail |
| POST | `/torrents/{id}/start` | Start |
| POST | `/torrents/{id}/stop` | Stop |
| POST | `/torrents/{id}/remove` | Remove |
| GET | `/events` | SSE stream |
| GET | `/static/*` | CSS and JS |

SSE event names: `stats`, `list`, `detail`, `removed`. Each `data` payload is an HTML fragment except `removed`, which is the torrent id.

## Project Structure

```
fluxo/
├── cmd/fluxo/          # Main application entry point
├── internal/
│   ├── config/         # Configuration (Cobra + Viper)
│   ├── server/         # HTTP server and listener
│   ├── webui/          # templ pages, SSE, form handlers
│   └── session/        # Torrent session manager and event bus
├── web/                # Tailwind/daisyUI build and embedded static files
├── go.mod
└── README.md
```

## License

MIT
