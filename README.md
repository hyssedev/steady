# Steady

Steady is a small self-hosted uptime monitor written in Go.


## Run

```sh
go run ./cmd/steady
```

Steady reads `config.yaml` from the working directory and creates `steady.db` when it does not exist.

## License

[MIT](LICENSE)
