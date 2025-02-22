
if ( (Read-Host "link up and build homie?[y/n]") -eq "y" ) {
    # build new droplet if needed
    pulumi refresh -C .\build\
    pulumi up -C .\build

    Write-Output "take that ip to cf homie"
    exit
}

# vars check
$strArr = (pulumi stack output -C .\build | Select-String "dropletIP") -split "\s+"
$DropletIP = $strArr[$strArr.Length - 1]
if ( -not $DropletIP || $DropletIP -eq "" ) {
    Write-Output "missing dropletIP output from pulumi"
    exit
}

Write-Output "connecting to droplet ip $DropletIP"
$BadChessPort = "8080"

# build go binary for target system
$os = $env:GOOS; $arch = $env:GOARCH
$env:GOOS = "linux"; $env:GOARCH = "amd64"

go build -ldflags '-s' -o .\remote\bad-chess-server ..\cmd\web\

$env:GOOS = $os; $env:GOARCH = $arch

# template Caddyfile
$CaddyTemplate = @"
{
    email badchessmgmt@gmail.com
}

http://bad-chess.com {
    respond /metrics/* "Not Permitted" 403
    reverse_proxy localhost:$BadChessPort
}
"@

#$Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False
#[System.IO.File]::WriteAllLines("./remote/Caddyfile", $CaddyTemplate, $Utf8NoBomEncoding)
$CaddyTemplate | Out-File -Encoding utf8 -FilePath ".\remote\Caddyfile" -Force

# copy binary & setup files to server
scp -r .\remote ("root@{0}:~/" -f $DropletIP)
ssh -t ("root@{0}" -f $DropletIP) "chmod +x ~/remote/server_setup.sh ~/remote/bad-chess-server && ~/remote/server_setup.sh"
