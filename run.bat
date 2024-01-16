@echo off

rem Build Go program
go build -o idie.exe idie.go

rem Check if the build was successful
if %errorlevel% neq 0 (
  echo Build failed
  exit /b 1
)

rem Run the compiled program
idie.exe