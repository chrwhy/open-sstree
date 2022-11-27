CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build  -o sstree_linux main.go

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o sstree_win.exe main.go

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o sstree_mac main.go

1. 中文搜索
2. 中文+全拼搜索
3. 中文+首字母
4. 全拼+中文搜索
5. 全拼搜索
6. 首字母搜索 


#open-sstree  
Open Search Suggestion Tree

搜索建议树
基于预设词条的搜索建议，支持中文，拼音，英文等混合搜索

原理
基于前缀树(字典树)实现，启动时将

/search?keyword=aaa&cate=bbb

命令行模式
go run main.go

web模式(默认开启8081端口)
go run main.go web