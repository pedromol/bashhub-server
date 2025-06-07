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

package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pedromol/bashhub-server/internal/db"
)

func getLog(logFile string) io.Writer {
	switch {
	case logFile == "/dev/null":
		return io.Discard
	case logFile != "":
		f, err := os.Create(logFile)
		if err != nil {
			log.Fatal(err)
		}
		return f
	default:
		return os.Stderr
	}
}

// loggerWithFormatterWriter instance a Logger middleware with the specified log format function.
func loggerWithFormatterWriter(logFile string, f gin.LogFormatter) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: f,
		Output:    getLog(logFile),
	})
}

// configure routes and middleware
func setupRouter(dbPath string, logFile string, registration bool) *gin.Engine {
	db.Init(dbPath)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(loggerWithFormatterWriter(logFile, func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[BASHHUB-SERVER] %v | %3d | %13v | %15s | %-7s  %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))

	// the jwt middleware
	key, _ := db.GetSecret()
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "bashhub-server zone",
		Key:         []byte(key),
		Timeout:     10000 * time.Hour,
		MaxRefresh:  10000 * time.Hour,
		IdentityKey: "username",
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"accessToken": token,
			})
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*db.User); ok {
				return jwt.MapClaims{
					"username":   v.Username,
					"systemName": v.SystemName,
					"user_id":    v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			var id uint
			switch claims["user_id"].(type) {
			case float64:
				id = uint(claims["user_id"].(float64))

			default:
				id = claims["user_id"].(uint)
			}
			return &db.User{
				Username:   claims["username"].(string),
				SystemName: claims["systemName"].(string),
				ID:         id,
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var user db.User

			if err := c.ShouldBind(&user); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			exists := user.UserExists()
			if exists != nil {
				user.SystemName, _ = user.UserGetSystemName()
				user.ID, _ = user.UserGetID()
				return &user, nil
			}
			fmt.Println("failed")

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			v, ok := data.(*db.User)
			exists, _ := v.UsernameExists()
			if ok && exists {
				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/api/v1/login", authMiddleware.LoginHandler)

	r.POST("/api/v1/user", func(c *gin.Context) {
		var user db.User
		if !registration {
			c.String(403, "Registration of new users is not allowed.")
			return
		}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if user.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email required"})
			return
		}
		if u, _ := user.UsernameExists(); u {
			c.String(409, "Username already taken")
			return
		}
		if e, _ := user.EmailExists(); e {
			c.String(409, "This email address is already registered.")
			return
		}
		user.UserCreate()
	})

	r.Use(authMiddleware.MiddlewareFunc())

	r.GET("/api/v1/command/:path", func(c *gin.Context) {
		var command db.Command
		var user db.User
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			command.User.ID = uint(claims["user_id"].(float64))

		default:
			command.User.ID = claims["user_id"].(uint)
		}

		if c.Param("path") == "search" {
			command.Limit = 100
			if c.Query("limit") != "" {
				if num, err := strconv.Atoi(c.Query("limit")); err != nil {
					command.Limit = 100
				} else {
					command.Limit = num
				}
			}
			command.Unique = false
			if c.Query("unique") == "true" {
				command.Unique = true
			}
			command.Path = c.Query("path")
			command.Query = c.Query("query")
			command.SystemName = c.Query("systemName")

			result, err := command.CommandGet()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if len(result) != 0 {
				c.IndentedJSON(http.StatusOK, result)
				return
			} else {
				c.JSON(http.StatusOK, gin.H{})
			}

		} else {
			command.Uuid = c.Param("path")
			result, err := command.CommandGetUUID()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			result.Username = user.Username
			c.IndentedJSON(http.StatusOK, result)
		}

	})

	r.POST("/api/v1/command", func(c *gin.Context) {
		var command db.Command
		if err := c.ShouldBindJSON(&command); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if command.ExitStatus != 0 && command.ExitStatus != 130 {
			return
		}
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			command.User.ID = uint(claims["user_id"].(float64))

		default:
			command.User.ID = claims["user_id"].(uint)
		}

		command.SystemName = claims["systemName"].(string)
		command.CommandInsert()
		c.AbortWithStatus(http.StatusOK)
	})

	r.DELETE("/api/v1/command/:uuid", func(c *gin.Context) {
		var command db.Command
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			command.User.ID = uint(claims["user_id"].(float64))

		default:
			command.User.ID = claims["user_id"].(uint)
		}
		command.Uuid = c.Param("uuid")
		command.CommandDelete()
		c.AbortWithStatus(http.StatusOK)

	})

	r.POST("/api/v1/system", func(c *gin.Context) {
		var system db.System
		err := c.Bind(&system)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			system.User.ID = uint(claims["user_id"].(float64))

		default:
			system.User.ID = claims["user_id"].(uint)
		}

		system.SystemInsert()
		c.AbortWithStatus(201)
	})

	r.GET("/api/v1/system", func(c *gin.Context) {
		var system db.System
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			system.User.ID = uint(claims["user_id"].(float64))

		default:
			system.User.ID = claims["user_id"].(uint)
		}
		mac := c.Query("mac")
		if mac == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		system.Mac = mac
		result, err := system.SystemGet()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, result)

	})

	r.PATCH("/api/v1/system/:mac", func(c *gin.Context) {
		var system db.System
		err := c.Bind(&system)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			system.User.ID = uint(claims["user_id"].(float64))

		default:
			system.User.ID = claims["user_id"].(uint)
		}
		system.Mac = c.Param("mac")
		system.SystemUpdate()
		c.AbortWithStatus(http.StatusOK)
	})

	r.GET("/api/v1/client-view/status", func(c *gin.Context) {
		var status db.Status
		claims := jwt.ExtractClaims(c)
		switch claims["user_id"].(type) {
		case float64:
			status.User.ID = uint(claims["user_id"].(float64))

		default:
			status.User.ID = claims["user_id"].(uint)
		}

		status.SessionName = c.Query("processId")
		t, err := strconv.Atoi(c.Query("startTime"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		status.SessionStartTime = int64(t)

		pid, err := strconv.Atoi(c.Query("processId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		status.ProcessID = pid

		result, err := status.StatusGet()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, result)
	})

	r.POST("/api/v1/import", func(c *gin.Context) {
		var imp db.Import
		if err := c.ShouldBindJSON(&imp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		claims := jwt.ExtractClaims(c)
		user := claims["username"].(string)
		imp.Username = user
		err = db.ImportCommands(imp)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatus(http.StatusOK)
	})

	return r
}

// Run starts server
func Run(dbFile string, logFile string, addr string, registration bool) {
	r := setupRouter(dbFile, logFile, registration)

	addr = strings.ReplaceAll(addr, "http://", "")
	err := r.Run(addr)

	if err != nil {
		fmt.Println("Error: \t", err)
	}

}
