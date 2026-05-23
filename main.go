package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"bd-ui/server"
)

//go:embed all:web-dist
var webFS embed.FS

func main() {
	host := flag.String("host", "", "监听地址 (默认 127.0.0.1)")
	port := flag.Int("port", 0, "监听端口 (默认 3000)")
	open := flag.Bool("open", false, "启动后打开浏览器")
	flag.Parse()

	if *host != "" {
		os.Setenv("HOST", *host)
	}
	if *port != 0 {
		os.Setenv("PORT", fmt.Sprintf("%d", *port))
	}

	config := server.GetConfig()

	webContent, _ := fs.Sub(webFS, "web-dist")

	sv := server.NewServer(config, webContent)
	if err := sv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	_ = open
}
