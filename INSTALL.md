# BIND DNS API - Installation Guide

## System Requirements

- Go 1.21+ (for building from source)
- BIND9 installed (for rndc operations)
- Linux with systemd (for service management)

## Quick Installation

### Option 1: Pre-built Binary

1. Copy the binary and configuration:
   ```bash
   sudo mkdir -p /opt/bind-dns-api/zones
   sudo cp bind-dns-api /opt/bind-dns-api/
   sudo cp config.json /opt/bind-dns-api/
   ```

2. Set permissions:
   ```bash
   sudo chown -R named:named /opt/bind-dns-api
   sudo chmod 755 /opt/bind-dns-api/bind-dns-api
   sudo chmod 644 /opt/bind-dns-api/config.json
   ```

3. Install systemd service:
   ```bash
   sudo cp bind-dns-api.service /etc/systemd/system/
   sudo systemctl daemon-reload
   ```

4. Enable and start:
   ```bash
   sudo systemctl enable bind-dns-api
   sudo systemctl start bind-dns-api
   sudo systemctl status bind-dns-api
   ```

### Option 2: Build from Source

1. Build the binary:
   ```bash
   go build -o bind-dns-api ./cmd/server
   ```

2. Follow steps from Option 1.

### Option 3: Development Mode

Run directly without installation:
```bash
go run ./cmd/server/main.go -config config.json
```

## Configuration

Edit `/opt/bind-dns-api/config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "bind": {
    "named_conf_path": "/etc/bind/named.conf",
    "zone_directory": "/opt/bind-dns-api/zones",
    "rndc_path": "/usr/sbin/rndc",
    "rndc_conf_path": "/etc/bind/rndc.conf"
  }
}
```

## BIND Integration

### Add zones to named.conf

Edit `/etc/bind/named.conf` and add:

```
include "/etc/bind/zones.conf";
```

Create `/etc/bind/zones.conf`:

```
zone "example.com" {
    type master;
    file "/opt/bind-dns-api/zones/example.com.zone";
};
```

Or use wildcards for dynamic zones:

```
zone "*" {
    type master;
    file "/opt/bind-dns-api/zones/%d.zone";
    allow-update { none; };
};
```

### Configure rndc

Ensure `/etc/bind/rndc.conf` exists:

```
default_key;
key "rndc-key" {
    algorithm hmac-sha256;
    secret "your-secret-key";
};
```

## Systemd Commands

| Command | Description |
|---------|-------------|
| `systemctl start bind-dns-api` | Start the service |
| `systemctl stop bind-dns-api` | Stop the service |
| `systemctl restart bind-dns-api` | Restart the service |
| `systemctl reload bind-dns-api` | Reload configuration |
| `systemctl status bind-dns-api` | Check service status |
| `systemctl enable bind-dns-api` | Enable on boot |
| `systemctl disable bind-dns-api` | Disable on boot |
| `journalctl -u bind-dns-api -f` | View logs |

## Firewall Configuration

Allow API access (adjust port if needed):

```bash
# UFW
sudo ufw allow 8080/tcp

# firewalld
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

## Testing

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Create domain
curl -X POST http://localhost:8080/api/v1/domains \
  -H "Content-Type: application/json" \
  -d '{"name": "example.com"}'

# List domains
curl http://localhost:8080/api/v1/domains
```

## Troubleshooting

### Check logs
```bash
journalctl -u bind-dns-api -n 50 --no-pager
```

### Verify BIND connectivity
```bash
sudo rndc status
```

### Check zone files
```bash
ls -la /opt/bind-dns-api/zones/
```

### Test configuration
```bash
/opt/bind-dns-api/bind-dns-api -config /opt/bind-dns-api/config.json
```

## Uninstall

```bash
sudo systemctl stop bind-dns-api
sudo systemctl disable bind-dns-api
sudo rm /etc/systemd/system/bind-dns-api.service
sudo systemctl daemon-reload
sudo rm -rf /opt/bind-dns-api
```
