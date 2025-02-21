# [Environment]::SetEnvironmentVariable("VariableName", "VariableValue", "User")

# build go binary for target system
$os = $env:GOOS; $arch = $env:GOARCH
$env:GOOS = "linux"; $env:GOARCH = "amd64"

go build -ldflags '-s' -o .\remote\bad-chess-server ..\cmd\web\

$env:GOOS = $os; $env:GOARCH = $arch

# copy binary & setup files to server
$ip = $env:BADCHESSDROPLETIP
scp -r ./remote ("root@{0}:~/" -f $ip)

ssh -t ("root@{0}:~/" -f $ip) "chmod +x ~/remote/server_setup.sh ~/remote/bad-chess-server && ~/remote/server_setup.sh"