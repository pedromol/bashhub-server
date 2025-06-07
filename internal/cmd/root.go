/*
 *
 * Copyright © 2020 nsherron90 <nsherron90@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pedromol/bashhub-server/internal/server"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	logFile      string
	dbPath       string
	addr         string
	registration bool
	rootCmd      = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Flags().Parse(args)
			startupMessage()
			server.Run(dbPath, logFile, addr, registration)
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
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", sqlitePath(), "db location (sqlite or postgres)")
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
	log.Println(banner)
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
