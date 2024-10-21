sudo apt-get install gcc-multilib
sudo apt-get install gcc-mingw-w64
sudo apt-get install golang

GOOS=windows GOARCH=386 CGO_ENABLED=1 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -tags noaudio