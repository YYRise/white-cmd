//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		configPath = flag.String("config", "./conf/whitelist.yaml", "白名单配置文件路径")
		cmd        = flag.String("cmd", "", "需要验证的shell命令")
	)
	flag.Parse()

	if *cmd == "" {
		fmt.Fprintln(os.Stderr, "错误: 必须指定--cmd参数")
		os.Exit(1)
	}

	validator, err := whitelist.NewValidator(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化验证器失败: %v\n", err)
		os.Exit(1)
	}

	allowed, err := validator.Validate(*cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "验证失败: %v\n", err)
		os.Exit(1)
	}

	if allowed {
		fmt.Println("命令符合白名单规则")
		os.Exit(0)
	}
	fmt.Println("命令不符合白名单规则")
	os.Exit(2)
}
