@echo off
echo ========================================
echo   Clearing Redis user_area cache
echo ========================================
echo.
redis-cli -h 134.175.206.158 -p 6379 -a aqi1015 DEL user_area
echo.
echo ========================================
echo   Cache cleared! Please restart Go service
echo ========================================
pause
