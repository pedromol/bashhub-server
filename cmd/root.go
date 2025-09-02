package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/pedromol/bashhub-server/internal"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	logFile      string
	dbPath       string
	addr         string
	registration bool
	traceProfile = os.Getenv("BH_SERVER_DEBUG_TRACE")
	cpuProfile   = os.Getenv("BH_SERVER_DEBUG_CPU")
	memProfile   = os.Getenv("BH_SERVER_DEBUG_MEM")
	rootCmd      = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Flags().Parse(args)
			checkBhEnv()
			startupMessage()
			if cpuProfile != "" || memProfile != "" || traceProfile != "" {
				profileInit()
			}
			internal.Run(dbPath, logFile, addr, registration)
		},
	}
)

// Execute runs root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "", `Set filepath for HTTP log. "" logs to stderr`)
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", postgresPath(), "PostgreSQL connection string")
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", listenAddr(), "Ip and port to listen and serve on")
	rootCmd.PersistentFlags().BoolVarP(&registration, "registration", "r", true, "Allow user registration")

}

// StartupMessage prints startup banner
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
                                                                                  
`, Version, addr, registration)
	color.HiGreen(banner)
	log.Printf("\nListening and serving HTTP on %v\n", addr)
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

func postgresPath() string {
	// Default PostgreSQL connection for local development
	return "postgres://postgres:password@localhost:5432/bashhub?sslmode=disable"
}


func checkBhEnv() {
	bhURL := os.Getenv("BH_URL")
	if strings.Contains(bhURL, "https://bashhub.com") {
		msg := fmt.Sprintf(`
WARNING: BH_URL is set to https://bashhub.com on this machine
If you will be running bashhub-client locally be sure to add
export BH_URL=%v to your .bashrc or .zshrc`, addr)
		fmt.Println(msg)
	}
}

func profileInit() {

	go func() {
		defer os.Exit(1)
		if traceProfile != "" {
			f, err := os.Create(traceProfile)
			if err != nil {
				log.Fatal("could not create trace profile: ", err)
			}
			defer f.Close()
			if err := trace.Start(f); err != nil {
				log.Fatal("could not start trace profile: ", err)
			}
			defer trace.Stop()
		}

		if cpuProfile != "" {
			f, err := os.Create(cpuProfile)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			defer pprof.StopCPUProfile()
		}

		defer func() {
			if memProfile != "" {
				mf, err := os.Create(memProfile)
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				defer mf.Close()
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(mf); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
			}
		}()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
	}()
}
