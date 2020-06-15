# filemanager
A Simple file upload, download server

# Usage

./filemanage 8090

## upload

`curl -T filename http://localhost:8090`
or
`curl -F 'file=@filename' http://localhost:8090`

## download

`wget http://localhost:8090/filename`

# Developer

`go build -v .`
