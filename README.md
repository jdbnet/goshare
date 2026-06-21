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

Download the pre-compiled binary:

```bash
wget https://apps.jdbnet.co.uk/goshare
chmod +x goshare
sudo mv goshare /usr/local/bin/
```

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
    dir: ./data
    max_upload_size: 104857600
```

## Run

After building, start the server with:

```bash
./goshare
```

The server will start on port 8080 (unless configured otherwise), and files will be stored in the `./data` directory relative to where it was executed.

## Running as a Service (Systemd)

We highly recommend running GoShare via Systemd so that it starts automatically on boot and runs continuously in the background.

1. Create a configuration directory and move your `config.yaml` there:
   ```bash
   sudo mkdir -p /etc/goshare
   sudo cp config.yaml /etc/goshare/config.yaml
   ```

2. Create a systemd service file at `/etc/systemd/system/goshare.service`:
   ```ini
   [Unit]
   Description=GoShare
   After=network.target

   [Service]
   Type=simple
   User=root
   ExecStart=/usr/local/bin/goshare --config /etc/goshare/config.yaml
   Restart=on-failure
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

3. Enable and start the service:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable goshare
   sudo systemctl start goshare
   ```