server:
	go build -o go-soundboard main.go

ui:
	yarn --cwd ./webinterface/ run build

run:
	go run main.go