ps -ef | grep sstree_linux | grep -v grep | cut -c 9-15 | xargs kill -s 9 && rm ./sstree_linux && mv ../sstree_linux ./ && nohup ./sstree_linux web &
