export GOPROXY=https://goproxy.cn
export GO111MODULE=on
export CGO_ENABLED=1
# export CXX=i686-w64-mingw32-g++
# export CC=i686-w64-mingw32-gcc

mkdir -p ./release-linux

function build {
    export GOOS=$1
    export GOARCH=$2
    
    if [ -d "./release-linux/$1-$2" ]; then rm -Rf "./release-linux/$1-$2"; fi

    mkdir -p "./release-linux/$1-$2"

    cp -r "./assets" "./release-linux/$1-$2/assets"
    mkdir -p  "./release-linux/$1-$2/conf"
    cp -r "./conf/setup.json" "./release-linux/$1-$2/conf/setup.json"

    go build -o "./release-linux/$1-$2/youcomic$3" main.go
}

function buildDocker {
    export GOOS=linux
    export GOARCH=amd64
    
    if [ -d "./release-linux/docker" ]; then rm -Rf "./release-linux/docker"; fi

    mkdir -p "./release-linux/docker"

    cp -r "./assets" "./release-linux/docker/assets"
    mkdir -p  "./release-linux/docker/conf"
    cp -r "./conf/setup.json" "./release-linux/docker/conf/setup.json"
    cp -r "./docker" "./release-linux"

    go build -o "./release-linux/docker/youcomic" main.go
}
echo "------------------------------ YouComic Builder --------------------------------"
echo "select build target with index,example: 1,2,3,4 "
echo "1. linux x64"
echo "2. linux x32"
echo "3. docker"

echo "make your select:"
read rawSelect
IFS=',' read -r -a selectArray <<< "$rawSelect"


for i in "${selectArray[@]}"
do
   case $i in
        1)
        build "linux" "amd64" ""
        ;;
        2)
        build "linux" "386" ""
        ;;
        3) buildDocker
        ;;
    esac
done