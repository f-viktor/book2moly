package main

import (
	"flag"
	"os"
)

type Config struct {
	HttpProxy string
}

type Arguments struct {
	Username  string
	Password  string
	TargetUrl string
}

var GlobalConfig Config

// parse command line arguments
func parseArgs() Arguments {

	//	config := flag.String("config", "config.json", "Path to config file")
	help := flag.Bool("h", false, "Display this help text")
	username := flag.String("u", "", "your moly.hu username (e-mail address)")
	password := flag.String("p", "", "your moly.hu password")
	targetUrl := flag.String("s", "https://www.goodreads.com/book/show/299215.The_Road_to_Serfdom", "source URL of the book you want to port to moly.hu")
	proxy := flag.String("proxy", "", "use proxy for debugging e.g: http://127.0.0.1:8000")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(2)
	}

	GlobalConfig.HttpProxy = *proxy

	args := Arguments{*username, *password, *targetUrl}
	return args
}
