
prepare-pi-dev:
		sudo apt install golang
		go mod download

run:
		go run *.go