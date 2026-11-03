# filedrop

A tiny file sharing server in Go with no dependencies. Upload a file, hand out its name, download it.

```sh
go run .
curl -F file=@report.pdf localhost:8080/upload
curl -OJ 'localhost:8080/download?name=report.pdf'
curl localhost:8080/files
```

## Configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `FILEDROP_DIR` | `./uploads` | where files are stored |
| `ADDR` | `:8080` | listen address |

Uploads are capped at 10 MiB.

## Tests

```sh
go test ./...
```
