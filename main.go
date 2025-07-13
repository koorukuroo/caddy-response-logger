package main

import (
    caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/koorukuroo/caddy-response-logger"
)

func main() {
    caddycmd.Main()
}
