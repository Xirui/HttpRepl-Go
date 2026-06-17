# HttpRepl-Go

A lightweight, cross-platform CLI REPL for testing REST APIs documented with OpenAPI/Swagger, written in Go.

## Overview

HttpRepl-Go is a Go port of [dotnet HttpRepl](https://github.com/dotnet/HttpRepl). Unlike the original, it compiles to a single binary with no runtime dependency — making it ideal for IoT, edge computing, CI/CD pipelines, and any environment where installing the .NET runtime is impractical.

## Features

- **OpenAPI tree navigation** — browse API endpoints using `ls` / `cd` / `tree`
- **HTTP method support** — `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`
- **Persistent custom headers** — set with `set header <name> <value>`, clear with `clear header <name>`
- **JSON pretty-printing** — auto-formatted response bodies
- **Response header display** — status and headers shown with every response
- **Customizable base address** — target any server via `--base-address` / `-b`
- **Configurable OpenAPI path** — specify the Swagger JSON endpoint with `--openapi` / `-o`
- **Start URL option** — jump directly into a path with `--start-url` / `-u`
- **Prompt with label** — current path shown in the interactive prompt

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+

### Build

```bash
git clone https://github.com/xirui/HttpRepl-Go.git
cd HttpRepl-Go
make install   # download dependencies
make build     # compile binary
```

The binary `httpreplgo` is produced in the current directory.

### Run directly

```bash
go run .
```

## Usage

```bash
httpreplgo [options]
```

### Options

| Flag                    | Alias | Default                   | Description                            |
| ----------------------- | ----- | ------------------------- | -------------------------------------- |
| `--base-address`        | `-b`  | `http://localhost:8080`   | Target server base address and port    |
| `--openapi`             | `-o`  | `/swagger/doc.json`       | Path to the OpenAPI/Swagger JSON       |
| `--start-url`           | `-u`  | `/api/v1`                 | Initial path to navigate to on startup |

### Commands

| Command                             | Description                            |
| ----------------------------------- | -------------------------------------- |
| `ls`                                | List child endpoints of current path   |
| `cd <path>`                         | Navigate to a path                     |
| `cd ..`                             | Go up one level                        |
| `cd`                                | Return to root (`/`)                   |
| `tree`                              | Display the full endpoint tree         |
| `get <id>`                          | Send GET request                       |
| `post <id>`                         | Send POST request                      |
| `put <id>`                          | Send PUT request                       |
| `delete <id>`                       | Send DELETE request                    |
| `patch <id>`                        | Send PATCH request                     |
| `head <id>`                         | Send HEAD request                      |
| `options <id>`                      | Send OPTIONS request                   |
| `set header <name> <value>`         | Set a custom header for all requests   |
| `clear header <name>`               | Clear a previously set custom header   |
| `help`                              | Display available commands             |
| `exit`                              | Exit the REPL                          |

### Example

```bash
$ httpreplgo -b https://api.example.com -o /v3/api-docs -u /api/v1
Using a base address of https://api.example.com
Checking https://api.example.com/v3/api-docs... Found
Parsing... Successful
- https://api.example.com
  - /api
    - /v1
      - get /api/v1/users
      - post /api/v1/users
      - get /api/v1/users/{id}
      - put /api/v1/users/{id}
      - delete /api/v1/users/{id}

https://api.example.com/api/v1> ls
.
..
users

https://api.example.com/api/v1> cd users

https://api.example.com/api/v1/users> get 42
https://api.example.com/api/v1/users/42
200 OK
content-type: application/json
...
{
  "id": 42,
  "name": "John Doe"
}
```

## Clean up

```bash
make clean
```

## License

[MIT](LICENSE)
