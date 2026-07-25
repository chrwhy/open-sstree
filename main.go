package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/chrwhy/open-sstree/tree"
)

func main() {
	dictDir := "."
	if dir := os.Getenv("SSTREE_DICT_DIR"); dir != "" {
		dictDir = dir
	}

	slog.Info("Initializing SSTree", "dictDir", dictDir)
	engine, err := sstree.New(dictDir)
	if err != nil {
		slog.Error("Failed to initialize SSTree", "error", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "web" {
		web(engine)
	} else {
		local(engine)
	}
}

func local(engine *sstree.SSTree) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Please input: ")
		keyword, _ := reader.ReadString('\n')
		if len(keyword) < 1 {
			continue
		}

		t0 := time.Now()
		candidates := engine.Search(sstree.DEFAULT_FOREST, keyword)
		t1 := time.Now()
		slog.Info("Search completed", "cost", t1.Sub(t0), "candidates", len(candidates))
		suggestions := sstree.XTraverse(candidates)
		slog.Info("Suggestions", "len", len(suggestions))
		sstree.PrintSuggestions(suggestions)
		slog.Info("Total", "cost", time.Since(t0))
	}
}

func web(engine *sstree.SSTree) {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/reload", func(c *gin.Context) {
		if err := engine.Reload(); err != nil {
			slog.Error("Reload failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": -1,
				"msg":  "reload failed: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
		})
	})

	r.GET("/search", func(c *gin.Context) {
		keyword, _ := c.GetQuery("keyword")
		keyword = strings.ToLower(keyword)
		cate, _ := c.GetQuery("cate")
		if len(cate) == 0 {
			cate = "default"
		}

		slog.Info("Search request", "keyword", keyword, "cate", cate)

		t0 := time.Now()
		result := engine.Search(cate, keyword)
		t1 := time.Now()
		slog.Info("Search completed", "cost", t1.Sub(t0), "totalRecords", len(result))

		// Cap traversal at 100 candidates for performance
		candidates := result
		if len(result) > 100 {
			candidates = result[0:100]
		}
		suggestions := sstree.XTraverse(candidates)
		slog.Info("Suggestions", "len", len(suggestions))

		// Extract text strings for API response
		resultTexts := make([]string, len(suggestions))
		for i, s := range suggestions {
			resultTexts[i] = s.Text
		}

		c.JSON(http.StatusOK, gin.H{
			"search_type": "",
			"result":      resultTexts,
		})
	})

	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		slog.Info("Received signal, shutting down", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		}
	}()

	slog.Info("Starting server", "addr", ":8081")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}
