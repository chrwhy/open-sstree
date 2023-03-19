package main

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"sstree/tree"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "web" {
		web()
	} else {
		local()
	}
}

func local() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please input: ")
		keyword, _ := reader.ReadString('\n')
		if len(keyword) < 1 {
			continue
		}

		t0 := time.Now()
		log.SetOutput(io.Discard)
		candidates := sstree.Search(sstree.DEFAULT_FOREST, keyword)
		t1 := time.Now()
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
		log.Println("Search cost:", t1.Sub(t0))
		suggestions := sstree.XTraverse(candidates)
		log.Println("Suggestions len:", len(suggestions))
		sstree.PrintSuggestions(suggestions)
		t2 := time.Now()
		log.Println("Total cost:", t2.Sub(t0))
	}
}

func web() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/reload", func(c *gin.Context) {
		sstree.MultiLoad()
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
		})
	})

	r.GET("/search", func(c *gin.Context) {
		log.Println("==========:", c.Query("keyword"))
		keyword, _ := c.GetQuery("keyword")
		keyword = strings.ToLower(keyword)
		cate, _ := c.GetQuery("cate")
		if len(cate) == 0 {
			cate = "default"
		}
		log.Println("keyword:", keyword)
		log.Println("cate:", cate)

		log.SetOutput(io.Discard)
		t0 := time.Now()
		result := sstree.Search(cate, keyword)
		t1 := time.Now()
		log.SetOutput(os.Stderr)
		log.Println("Search cost:", t1.Sub(t0))
		log.Println("Total records: ", len(result))
		suggestions := make([]string, 0)
		if len(result) > 100 {
			suggestions = sstree.XTraverse(result[0:100])
			log.Println("Suggestions len:", len(suggestions))
		} else {
			suggestions = sstree.XTraverse(result)
			log.Println("Suggestions len:", len(suggestions))
		}

		if len(result) > 100 {
			c.JSON(http.StatusOK, gin.H{
				"search_type": "",
				"result":      suggestions,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"search_type": "",
				"result":      suggestions,
			})
		}
	})

	r.Run(":8081")
}
