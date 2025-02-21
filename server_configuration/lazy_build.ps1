# vars check
$DropletIP = $env:BADCHESSDROPLETIP
if ( -not $DropletIP || $DropletIP -eq "" ) {
    Write-Output '$env:BADCHESSDROPLETIP environment variable missing
    set with [Environment]::SetEnvironmentVariable("VariableName", "VariableValue", "User")'
    exit
}

$BadChessPort = "8080"

# build go binary for target system
$os = $env:GOOS; $arch = $env:GOARCH
$env:GOOS = "linux"; $env:GOARCH = "amd64"

go build -ldflags '-s' -o .\remote\bad-chess-server ..\cmd\web\

$env:GOOS = $os; $env:GOARCH = $arch

# template Caddyfile
$CaddyTemplate = @"
http://$DropletIP {
        respond /metrics/* "Not Permitted" 403
        reverse_proxy localhost:$BadChessPort
}
"@

$CaddyTemplate | Out-File -FilePath "./remote/Caddyfile" -Force

# copy binary & setup files to server
scp -r ./remote ("root@{0}:~/" -f $DropletIP)
ssh -t ("root@{0}" -f $DropletIP) "chmod +x ~/remote/server_setup.sh ~/remote/bad-chess-server && ~/remote/server_setup.sh"
