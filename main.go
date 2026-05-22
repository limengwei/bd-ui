package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"bd-ui/server"
)

//go:embed all:web-dist
var webFS embed.FS

func main() {
	host := flag.String("host", "", "监听地址 (默认 127.0.0.1)")
	port := flag.Int("port", 0, "监听端口 (默认 3000)")
	dir := flag.String("dir", "", "Beads 项目目录 (默认当前目录)")
	open := flag.Bool("open", false, "启动后打开浏览器")
	flag.Parse()

	if *host != "" {
		os.Setenv("HOST", *host)
	}
	if *port != 0 {
		os.Setenv("PORT", fmt.Sprintf("%d", *port))
	}

	if *dir != "" {
		absDir, err := filepath.Abs(*dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无效目录: %s: %v\n", *dir, err)
			os.Exit(1)
		}
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "目录不存在: %s\n", absDir)
			os.Exit(1)
		}
		if err := os.Chdir(absDir); err != nil {
			fmt.Fprintf(os.Stderr, "无法切换到目录: %s: %v\n", absDir, err)
			os.Exit(1)
		}
	}

	config := server.GetConfig()

	fmt.Printf("工作目录: %s\n", config.RootDir)

	dbResult := server.ResolveWorkspaceDatabase(config.RootDir)
	fmt.Printf("数据库路径: %s (来源: %s, 存在: %v)\n", dbResult.Path, dbResult.Source, dbResult.Exists)
	if !dbResult.Exists {
		fmt.Fprintf(os.Stderr, "\n警告: 未找到 Beads 数据库！\n")
		fmt.Fprintf(os.Stderr, "请确保在包含 .beads/ 目录的项目中运行，或使用 --dir 指定项目路径。\n")
		fmt.Fprintf(os.Stderr, "示例: bd-ui --dir /path/to/your/project\n")
	}

	webContent, _ := fs.Sub(webFS, "web-dist")

	sv := server.NewServer(config, webContent)
	if err := sv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	_ = open
}
