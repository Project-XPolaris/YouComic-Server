
# init 
$Env:GO111MODULE = "on"
$Env:GOPROXY = "https://goproxy.cn,direct"
$Env:CGO_ENABLED = "1"
Remove-Item -Path "./release" -Recurse -Force -ErrorAction Ignore | Out-Null
New-Item -Path "./release" -Force -ItemType Directory | Out-Null
function GetBinaryExtension {
    param (
        [string]$OS
    )
    if ($OS.Contains("windows")) {
        return ".exe"
    }
    return ""
}
function GetCompressOutput {
    param ([string]$OS,[string]$Arch)
    switch ($OS) {
        "windows" { 
            Compress-Archive -Path "release/$($OS)-$($Arch)/*" -DestinationPath "release/$($OS)-$($Arch).zip"
         }
        Default {
            tar --strip-components 2 -czvf "release/$($OS)-$($Arch).tar.gz" -c "release/$($OS)-$($Arch)" | Out-Null
        }
    }
}
function BuildBinary {
    param (
        [string]$OS, [string]$Arch
    )
    Write-Host "--------------------- Build $($OS) $($Arch) Binary --------------------------"
    # prepare output directory
    Write-Host "create directory..." -NoNewline
    $outputDirectory = "./release/$($OS)-$($Arch)"
    New-Item -Path $outputDirectory -ItemType Directory | Out-Null
    New-Item -Path "$($outputDirectory)/conf" -ItemType Directory | Out-Null
    Write-Host "done" -ForegroundColor Green
    # copy resource file
    Write-Host "copy resource files..."  -NoNewline
    Copy-Item -Path "./assets" -Destination "$($outputDirectory)" -Recurse | Out-Null
    Copy-Item -Path "./conf/setup.json" -Destination "$($outputDirectory)/conf/setup.json" | Out-Null
    Write-Host "done" -ForegroundColor Green

    # build binary
    Write-Host "build binary..."  -NoNewline 
    $Env:GOOS = "$($OS)"
    $Env:GOARCH = "$($Arch)"
    $extension = GetBinaryExtension -OS $OS
    go build -o "$($outputDirectory)/youcomic$($extension)" main.go
    Write-Host "done" -ForegroundColor Green
    
    Write-Host "compress files..."  -NoNewline 
    GetCompressOutput -OS $OS -Arch $Arch
    Write-Host "done" -ForegroundColor Green
}

Write-Host "----------------------- YouComic Builder -----------------------------"
Write-Host 'select target with number,example 1,2,3,4'
Write-Host '1. windows x64
2. windows x32
3. linux x64
4. linux x32
5. darwin x64
6. darwin x32
7 freebsd x64
8 freebsd x32'

$selectString = Read-Host "your select:"

$options = $selectString.Split(",")

foreach ($option in $options) {
    switch ($option) {
        1 { 
            BuildBinary -OS "windows" -Arch "amd64"
        }
        2 {
            BuildBinary -OS "windows" -Arch "386"
        }
        3 {
            BuildBinary -OS "linux" -Arch "amd64"
        }
        4 {
            BuildBinary -OS "linux" -Arch "386"
        }
        5 {
            BuildBinary -OS "darwin" -Arch "amd64"
        }
        6 {
            BuildBinary -OS "darwin" -Arch "386"
        }
        7 {
            BuildBinary -OS "freebsd" -Arch "amd64"
        }
        8 {
            BuildBinary -OS "freebsd" -Arch "386"
        }
    }
}