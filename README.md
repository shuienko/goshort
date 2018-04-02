# goshort
----

## Description

*goshort* is URL Shortener that supports custom URL Schemes.

See API description below.

## Build

- Install `dep` https://github.com/golang/dep#installation
- Get dependencies
```bash
dep ensure 
```
- Build binary
```bash
go build -o goshort
```

## Usage

```
Usage of goshort:
  -dbbucket bucket
        DB bucket name (default "goshort")
  -dbpath string
        Path to the database file (default "goshort.db")
  -listen string
        Listen interface (default "127.0.0.1:8080")
  -salt string
        Hash salt string (default "tFZ2cQ7U8OQlSWOyZIoFdRusvkFvJh3A")
  -scheme string
        URL scheme http or https (default "http")
```

## API

Request body format: `{ "url": "http://exmple.com" }`

### Operations

| Endpoint   | Method   | Description                                                    |
| ---------- | -------- | -------------------------------------------------------------- |
| /{id}      | GET      | Will redirect you to the corresponding Long URL with 301 code. |
| /v1/short  | POST     | Will return short URL in the response body.                    |

### Examples

#### Short request

```
$ curl -d'{"url":"https://google.com"}' -XPOST http://s.shov.cloud/v1/short
{"shortURL":"http://s.shov.cloud/xqJW9dZ"}
```

#### Redirect

```
$ curl http://s.shov.cloud/xqJW9dZ
<a href="https://google.com">Moved Permanently</a>.
```