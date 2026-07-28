# TLS certificates for nginx (secure compose)

Place `tls.crt` and `tls.key` here for local secure overlay testing, or set `NGINX_TLS_CERT` / `NGINX_TLS_KEY` to mounted paths from your PKI.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt -subj '/CN=localhost'
```

Do not commit production private keys.
