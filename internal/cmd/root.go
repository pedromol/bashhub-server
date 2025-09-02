package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pedromol/bashhub-server/internal/server"
)

var (
	logFile      = flag.String("log", "", "log file location")
	dbPath       = flag.String("db", sqlitePath(), "db location (sqlite or postgres)")
	addr         = flag.String("addr", listenAddr(), "Ip and port to listen and serve on")
	registration = flag.Bool("registration", true, "Allow user registration")
	showVersion  = flag.Bool("version", false, "Show version information")
	GitCommit    string
	BuildDate    string
	Version      string
)

func Execute() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0)
	}

	startupMessage()
	server.Run(*dbPath, *logFile, *addr, *registration)
}
func startupMessage() {
	banner := fmt.Sprintf(`
 _               _     _           _
| |             | |   | |         | |		version: %v
| |__   __ _ ___| |__ | |__  _   _| |		address: %v
| '_ \ / _' / __| '_ \| '_ \| | | | '_ \	registration: %v
| |_) | (_| \__ \ | | | | | | |_| | |_) |
|_.__/ \__,_|___/_| |_|_| |_|\__,_|_.__/
 ___  ___ _ ____   _____ _ __
/ __|/ _ \ '__\ \ / / _ \ '__|
\__ \  __/ |   \ V /  __/ |
|___/\___|_|    \_/ \___|_|
`, Version, *addr, *registration)
	log.Println(banner)
	log.Printf("\nListening and serving HTTP on %v\n", *addr)
}
func listenAddr() string {
	var a string
	if os.Getenv("BH_SERVER_URL") != "" {
		a = os.Getenv("BH_SERVER_URL")
		return a
	}
	a = "http://0.0.0.0:8080"
	return a
}
func sqlitePath() string {
	dbFile := "data.db"
	f := filepath.Join(appDir(), dbFile)
	return f
}
func appDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	ch := filepath.Join(cfgDir, "bashhub-server")
	err = os.MkdirAll(ch, 0755)
	if err != nil {
		log.Fatal(err)
	}
	return ch
}
