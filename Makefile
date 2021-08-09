# Binary name
BINARY= Mule
VERSION = 0.3.3beta
# Builds the project
build:
		go build -ldflags "-s -w" -o ${BINARY} ./main.go
# Installs our project: copies binaries
install:
		go install
release-upx:
		# Clean
		#go clean
		rm -rf *.gz
		# Build for mac
		go build -ldflags "-s -w" -o ./bin/Mule-mac64-${VERSION} ./main.go
		upx -2 ./bin/Mule-mac64-${VERSION}
		# Build for linux
		#go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/Mule-linux64-${VERSION} ./main.go
		upx -2 ./bin/Mule-linux64-${VERSION}
		#go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o ./bin/Mule-linux32-${VERSION} ./main.go
		upx -2 ./bin/Mule-linux32-${VERSION}
		# Build for win
		#go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/Mule-win64-${VERSION}.exe  ./main.go
		upx -2 ./bin/Mule-win64-${VERSION}.exe
		#go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o ./bin/Mule-win32-${VERSION}.exe ./main.go
		upx -2 ./bin/Mule-win32-${VERSION}.exe
		cp ./ReadMe.md ./bin/
		cp -r ./Data ./bin/
		#compress
		tar cvf Mule-${VERSION}.tar.gz bin/*
# Cleans our projects: deletes binaries

release:
# Clean
		#go clean
		rm -rf *.gz
		# Build for mac
		go build -ldflags "-s -w" -o ./bin/Mule-mac64-${VERSION} ./main.go
		# Build for linux
		#go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/Mule-linux64-${VERSION} ./main.go
		#go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o ./bin/Mule-linux32-${VERSION} ./main.go
		# Build for win
		#go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/Mule-win64-${VERSION}.exe  ./main.go
		#go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o ./bin/Mule-win32-${VERSION}.exe ./main.go
		#compress
		tar cvf Mule-${VERSION}.tar.gz bin/*

clean:
		go clean

.PHONY:  clean build