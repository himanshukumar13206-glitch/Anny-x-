/*
 * ● ArcMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 Team Arc
 */

package main

/*
#cgo CFLAGS: -I../../
#cgo linux LDFLAGS: -L ../../ -lntgcalls -lm -lz
#cgo darwin LDFLAGS: -L ../../ -lntgcalls -lc++ -lz -lbz2 -liconv -framework AVFoundation -framework AudioToolbox -framework CoreAudio -framework QuartzCore -framework CoreMedia -framework VideoToolbox -framework AppKit -framework Metal -framework MetalKit -framework OpenGL -framework IOSurface -framework ScreenCaptureKit

// Currently is supported only dynamically linked library on Windows due to
// https://github.com/golang/go/issues/63903
#cgo windows LDFLAGS: -L../../ -lntgcalls
#include "ntgcalls/ntgcalls.h"
#include "glibc_compatibility.h"
*/
import "C"

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/Laky-64/gologging"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/modules"
	"main/internal/platforms"
)

func main() {
	initLogger()
	defer config.CloseLogging()

	shutdownPlatforms, err := platforms.Init()
	if err != nil {
		gologging.Fatal("Failed to initialize platforms: " + err.Error())
	}
	defer shutdownPlatforms()

	checkFFmpegAndFFprobe()

	if err := refreshDirs(); err != nil {
		gologging.Fatal("Failed to refresh directories: " + err.Error())
	}

	gologging.Debug("Initializing MongoDB...")

	closeDB, err := database.Init(config.MongoURI)
	if err != nil {
		gologging.Fatal("Failed to initialize database: " + err.Error())
	}
	defer closeDB()

	gologging.Info("Database connected successfully")

	if err := locales.Load(); err != nil {
		gologging.Fatal("Failed to load locales: " + err.Error())
	}

	gologging.Debug("Initializing clients...")

	shutdownCore, err := core.Init()
	if err != nil {
		gologging.Fatal("Failed to initialize core: " + err.Error())
	}
	defer shutdownCore()

	core.GetAssistantIndexFunc = database.AssistantIndex
	core.F = modules.F

	if err := database.RebalanceAssistantIndexes(core.Assistants.Count()); err != nil {
		gologging.Fatal("Failed to rebalance Assistants: " + err.Error())
	}

	modules.Init(core.Bot, core.Assistants)

	// --- MODIFIED: Start HTTP server with health check for Render ---
	startHTTPServer()

	core.Bot.Idle()
}

// startHTTPServer now includes a health check endpoint and uses Render's $PORT
func startHTTPServer() {
	go func() {
		// Determine the port: first try Render's PORT env, then config.Port, then default 8080
		port := os.Getenv("PORT")
		if port == "" {
			port = config.Port
			if port == "" {
				port = "8080"
			}
		}
		addr := "0.0.0.0:" + port

		// Add a health check endpoint for Render (and any other monitoring)
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Optional: add a root endpoint to avoid 404 noise
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ArcMusic Bot is running"))
		})

		gologging.Info("Starting HTTP server on " + addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			gologging.Error("HTTP server error: " + err.Error())
		}
	}()
}

func initLogger() {
	gologging.SetLevel(gologging.DebugLevel)
	gologging.SetOutput(config.LogWriter)

	l := gologging.GetLogger("ntgcalls")
	l.SetLevel(gologging.ErrorLevel)
	l.SetOutput(config.LogWriter)

	l = gologging.GetLogger("webrtc")
	l.SetLevel(gologging.ErrorLevel)
	l.SetOutput(config.LogWriter)

	gologging.GetLogger("Database").SetOutput(config.LogWriter)
}

func refreshDirs() error {
	dirs := []string{
		"./cache",
		"./downloads",
	}

	for _, dir := range dirs {

		if err := os.RemoveAll(dir); err != nil {
			return err
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
