# Gogunzy - ODK-X helper to support newer Android versions

Gogunzy is a lightweight Go-based proxy server designed to sit in front of the ODK-X Sync Endpoint to transparently handle and decompress incoming `gzip`-encoded HTTP request bodies. This is especially useful when dealing with clients (like the ODK-X Android tools) on newer android devices (e.g Android 14) that send gzipped requests, which may not be natively supported or expected by the sync backend yet.

This will cause an error where Sync Endpoint will silently store the gzipped files directly, which causes a failure when the device wants to synchronize with the endpoint at a later stage (because the device will then get a binary response in lieu of a properly formatted json document).

## Features

- Automatically decompresses gzipped request bodies.
- Forwards all headers (including `Authorization`) to preserve authentication.
- Does not lowercase HTTP headers as some other proxies do (while it certainly is more pleasing, it makes ODK-X break since it unfortunately validates on header case-sensitively)
- Adds an 'Accept-Encoding: identity' header to suggest the backend should avoid gzip. Note that this is set explicitly, rather than leaving the header empty, though the backend's behavior is not verified or enforced.
- Includes a `/gogunzy-health` endpoint to verify status


## Use Case

In Docker environments, especially when using NGINX as a reverse proxy, gzipped POST/PUT/PATCH requests can cause compatibility issues. Gogunzy mitigates this by sitting between NGINX and the ODK-X Sync Endpoint, handling decompression and header sanitization transparently.

## Deployment

### Using Pre-built Docker Image (Recommended)

The easiest way to deploy Gogunzy is to use the pre-built Docker image from GitHub Container Registry:

```bash
docker pull ghcr.io/hellosapiens/odkx-gogunzy:latest
```

You can then reference this image in your Docker Compose file as shown below.

### Docker Compose Example

```yaml
gogunzy:
  image: ghcr.io/hellosapiens/odkx-gogunzy:latest  # Using the pre-built image from GitHub Container Registry
  networks:
    - sync-network
  ports:
    - "8000:8000"
  restart: unless-stopped
```

### Building from Source (Alternative)

If you prefer to build the image yourself, you can clone this repository and build it locally:

```bash
git clone https://github.com/HelloSapiens/ODKX-Gogunzy.git
cd ODKX-Gogunzy
docker build -t gogunzy .
```

### NGINX Configuration
You need to update the nginx configuration, stored in the file: `/conf/nginx/sync-endpoint-locations.conf` to handle all /odktables/ requests off to the proxy for gzip inflation before they hit the sync endpoint.

```nginx
location ^~ /odktables/ {
    proxy_pass http://gogunzy:8000/odktables/; # passing to gogunzy instead of directly to sync:8080
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Port $server_port;
    proxy_set_header Host $host:$server_port;
    proxy_redirect default;
}

location /gogunzy-health {
    proxy_pass http://gogunzy:8000/gogunzy-health;
}
```

## License

MIT
