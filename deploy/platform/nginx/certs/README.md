# TLS certificates for unified platform nginx

Place `tls.crt` and `tls.key` here for local unified-edge testing, or set `UNIFIED_NGINX_TLS_CERT` / `UNIFIED_NGINX_TLS_KEY` (or `VEIL_EDGE_TLS_CERT` / `VEIL_EDGE_TLS_KEY`) to mounted paths from your PKI.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt -subj '/CN=localhost'
```

Do not commit production private keys. See [docs/architecture/platform-unified-access.md](../../../docs/architecture/platform-unified-access.md).
