CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build  -o sstree_linux main.go && scp ./sstree_linux root@111.230.143.95:/home/stephen/
