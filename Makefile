.SILENT:

build_arm:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o skin-downloader -v main.go