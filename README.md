<div align="center">
  <img src="web/templates/icon.png" alt="GoShare" width="128" />
  
  # GoShare

GoShare is a lightweight file-sharing web application written in Go.

</div>

## Features

- Upload, download, and delete files
- Password protection for uploads
- File expiration times
- Embedded web UI
- Single binary deployment
- Zero external dependencies
- Clean, modern interface

## Build

To build the project, run the following command in the root directory:

```bash
go build -o goshare .
```

This will create an executable named `goshare`.

## Installation

Install from the Aptuary APT repository:

```bash
curl -fsSL https://apt.jdbnet.co.uk/install/stable.sh | sudo bash
sudo apt install goshare
```

The `.deb` installs the binary, a systemd unit, and an example config at `/etc/goshare/config.yaml`.

## Configuration

By default, GoShare loads `/etc/goshare/config.yaml` when present. If the file is missing, built-in defaults apply (and can be overridden with environment variables).

You can specify a custom configuration file path using the `-config` flag:

```bash
./goshare -config ./config.yaml
```

Environment variables (override file values):

| Variable | Description |
|----------|-------------|
| `GOSHARE_SERVER_PORT` | HTTP port (default `8080`) |
| `GOSHARE_STORAGE_DIR` | Directory for uploaded files |
| `GOSHARE_MAX_UPLOAD_SIZE` | Max upload size in bytes |
| `GOSHARE_AUTH_ENABLED` | `true` or `false` |
| `GOSHARE_AUTH_USERNAME` | Upload auth username |
| `GOSHARE_AUTH_PASSWORD` | Upload auth password |

When auth is enabled, anyone with a file's share link can download that file without logging in. Share links look like `https://your-host/d/{token}`. Use **Copy link** in the file list to generate them; keep links private if the file is sensitive.

### Default Configuration

A sample default configuration:

```yaml
server:
    port: 8080
auth:
    enabled: false
    username: admin
    password: password
storage:
    dir: /var/lib/goshare/data
    max_upload_size: 104857600
```

## Run

After building, start the server with:

```bash
./goshare
```

The server will start on port 8080 (unless configured otherwise). When installed via the `.deb`, files are stored in `/var/lib/goshare/data`.

## Running as a Service (Systemd)

The `.deb` package installs `/lib/systemd/system/goshare.service` and creates a `goshare` system user with `/var/lib/goshare` for data storage. Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable goshare
sudo systemctl start goshare
```

### Manual setup

If you built from source instead of installing the package:

1. Create a configuration directory and move your `config.yaml` there:
   ```bash
   sudo mkdir -p /etc/goshare
   sudo cp config.yaml /etc/goshare/config.yaml
   ```

2. Copy the systemd unit from `docs/goshare.service` to `/etc/systemd/system/goshare.service`, create the `goshare` user, and set up data storage:
   ```bash
   sudo useradd -r -s /usr/sbin/nologin goshare || true
   sudo mkdir -p /var/lib/goshare/data
   sudo chown -R goshare:goshare /var/lib/goshare
   sudo cp docs/goshare.service /etc/systemd/system/
   ```

3. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable goshare
   sudo systemctl start goshare
   ```

## Docker

Images are published to `ghcr.io/jdbnet/goshare` on each release. The container uses the same binary as the `.deb` / GitHub release and configures itself via environment variables (no `config.yaml` required).

```bash
docker run -d --name goshare \
  -p 8080:8080 \
  -v goshare-data:/var/lib/goshare/data \
  -e GOSHARE_AUTH_ENABLED=true \
  -e GOSHARE_AUTH_USERNAME=admin \
  -e GOSHARE_AUTH_PASSWORD=changeme \
  ghcr.io/jdbnet/goshare:latest
```

Local image build (after `go build -o goshare .` or copying a release binary to `./goshare`):

```bash
docker build -t goshare:local .
```