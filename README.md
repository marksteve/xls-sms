# xlsx-sms

## Config

Create a `conf.toml` file:

```toml
[chikka]
client-id = "CHIKKA_CLIENT_ID"
secret-key = "CHIKKA_SECRET_KEY"
shortcode = "CHIKKA_SHORTCODE"
```

## Deploy

### Docker

```bash
docker-compose build
docker-compose up -d
```

### If you have a working Go environment...

```bash
go get
go run main.go
```

## License
http://marksteve.mit-license.org
