#GOOS=linux GOARCH=mips GOMIPS=softfloat go build -ldflags "-s -w" -o goredsocks
cd main && GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "-s -w" -o ../go-tproxy2socks

cd .. && upx --best go-tproxy2socks

#scp go-tproxy2socks root@10.10.1.1:/usr/bin
#scp monitor.sh root@10.10.1.1:/root/



