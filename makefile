.PHONY: all

run:
	@go run *.go --templatedir=./template/ --resultdir=./tmp/ --log-level=debug

all:
	@mkdir -p bin/
	@bash --norc -i ./scripts/build.sh

linux:
	@mkdir -p bin/
	@export GOOS=linux && export GOARCH=amd64 && bash --norc -i ./scripts/build.sh

deploy: linux
	@echo "--> Uploading..."
	scp -P 3389 contrib/init/peskar-index.default leo@paradev.ru:/etc/default/peskar-index
	scp -P 3389 contrib/init/peskar-index leo@paradev.ru:/etc/init.d/peskar-index
	scp -P 3389 bin/peskar-index leo@paradev.ru:/opt/peskar/peskar-index_new
	scp -P 3389 template/movie.html leo@paradev.ru:/opt/peskar/template/movie.html
	@echo "--> Restarting..."
	ssh -p 3389 leo@paradev.ru service peskar-index stop
	ssh -p 3389 leo@paradev.ru rm /opt/peskar/peskar-index
	ssh -p 3389 leo@paradev.ru mv /opt/peskar/peskar-index_new /opt/peskar/peskar-index
	ssh -p 3389 leo@paradev.ru service peskar-index start
	@echo "--> Getting last logs..."
	@ssh -p 3389 leo@paradev.ru tail -n 25 /var/log/peskar-index.log

logs:
	@ssh -p 3389 leo@paradev.ru tail -n 100 /var/log/peskar-index.log
