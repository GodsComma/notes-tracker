.PHONY: dev build clean

dev:
	@trap 'kill 0' EXIT; go run -tags fts5 ./server & cd client && npm run dev

build:
	cd client && npm run build
	go build -tags fts5 -o bin/notes-server ./server

clean:
	rm -rf bin/ client/dist/
