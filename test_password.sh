#!/usr/bin/expect -f

# Test script for interactive password input
spawn ./afe user create --name "Test User" --email "test@example.com"

expect "Enter password:"
send "test123\r"

expect "Confirm password:"
send "test123\r"

expect eof