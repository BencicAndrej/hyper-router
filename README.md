# Hyper Router

Hyper Router is a custom HTTP request router for golang.

The Hyper Router was heavily inspired by [julienschmidt/httprouter](https://github.com/julienschmit/httprouter).

## Features

- Replacement for the default `http.ServeMux` with a more flexible and faster routing definitions.
- Each request is extended with the `context.Context ` parameter for passing the request scoped data.
- A simple and elegant middleware system using the `hyper.MiddlewareStack`

## Usage

```go
package main

import (
    "github.com/bencicandrej/hyper-router"
)

func main() {
    router := hyper.Router{}
}
```