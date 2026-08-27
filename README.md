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

By default, GoShare looks for its configuration file at `/etc/goshare/config.yaml`. If the file does not exist, it will attempt to generate a default one.

You can specify a custom configuration file path using the `-config` flag:

```bash
./goshare -config ./config.yaml
```

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