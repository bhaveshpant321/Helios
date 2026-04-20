@echo off
echo Starting Helios UI Server...
echo.
echo UI will be available at: http://localhost:3000
echo Press Ctrl+C to stop
echo.

cd /d "%~dp0"
python -m http.server 3000