server:
	go build -o go-soundboard *.go

ui:
	yarn --cwd ./webinterface/ run build

dev:
	go run *.go || yarn --cwd ./webinterface/ run dev