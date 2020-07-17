


# init 
$Env:GO111MODULE = "on"
$Env:GOPROXY = "https://goproxy.cn,direct"
$Env:CGO_ENABLED = "1"
Remove-Item -Path "./release" -Recurse -Force -ErrorAction Ignore | Out-Null
New-Item -Path "./release" -Force -ItemType Directory | Out-Null

function GetCurrentLocationInWSL {
    $rawPath = Get-Location
    $projectPath = $rawPath -replace("\\","\\")
    return wsl wslpath -a $projectPath
}

# prepare wsl
wsl -e cd "$(GetCurrentLocationInWSL)"
wsl -e export GOPROXY=https://goproxy.cn
wsl -e export GO111MODULE=on
wsl -e export CGO_ENABLED=1


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
    param ([string]$OS, [string]$Arch)
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

function BuildLinux {
    param ([string]$Arch)
    
    Write-Host "--------------------- Build Linux $($Arch) Binary --------------------------"
    # prepare output directory
    Write-Host "create directory..." -NoNewline
    $outputDirectory = "./release/linux-$($Arch)"
    New-Item -Path $outputDirectory -ItemType Directory | Out-Null
    New-Item -Path "$($outputDirectory)/conf" -ItemType Directory | Out-Null
    Write-Host "done" -ForegroundColor Green
    # copy resource file
    Write-Host "copy resource files..."  -NoNewline
    Copy-Item -Path "./assets" -Destination "$($outputDirectory)" -Recurse | Out-Null
    Copy-Item -Path "./conf/setup.json" -Destination "$($outputDirectory)/conf/setup.json" | Out-Null
    Write-Host "done" -ForegroundColor Green
    
    Write-Host "build binary..."  -NoNewline 
    wsl go build -o "$($outputDirectory)/youcomic" main.go
    Write-Host "done" -ForegroundColor Green

    Write-Host "compress files..."  -NoNewline 
    wsl tar --strip-components 2 -czvf "release/linux-$($Arch).tar.gz" -c "release/linux-$($Arch)"
    Write-Host "done" -ForegroundColor Green
    
}

function BuildDocker{
    Write-Host "--------------------- Build Linux $($Arch) Binary --------------------------"
    # prepare output directory
    Write-Host "create directory..." -NoNewline
    $outputDirectory = "./release/docker"
    New-Item -Path $outputDirectory -ItemType Directory | Out-Null
    New-Item -Path "$($outputDirectory)/conf" -ItemType Directory | Out-Null
    Write-Host "done" -ForegroundColor Green
    # copy resource file
    Write-Host "copy resource files..."  -NoNewline
    Copy-Item -Path "./assets" -Destination "$($outputDirectory)" -Recurse | Out-Null
    Copy-Item -Path "./conf/setup.json" -Destination "$($outputDirectory)/conf/setup.json" | Out-Null
    Write-Host "done" -ForegroundColor Green
    Copy-Item -Path "./docker/*" -Destination $outputDirectory -Recurse | Out-Null

    Write-Host "build binary..."  -NoNewline
    wsl go build -o "$($outputDirectory)/youcomic" main.go
    Write-Host "done" -ForegroundColor Green

#    Write-Host "compress files..."  -NoNewline
#    wsl tar --strip-components 2 -czvf "release/docker.tar.gz" -c "release/docker"
#    Write-Host "done" -ForegroundColor Green
}

Write-Host "----------------------- YouComic Builder -----------------------------"
Write-Host 'select target with number,example 1,2,3,4'
Write-Host '1. windows x64
2. windows x32
3. linux x64
4. linux x32
5. docker'

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
            BuildLinux -Arch "amd64"
        }
        4 {
            BuildLinux -Arch "386"
        }
        5 {
            BuildDocker
        }
    }
}