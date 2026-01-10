package common

import "syscall"

func GetEnv(key, fallback string) string {
	if val, ok := syscall.Getenv(key); ok {
		return val
	}
	
	return fallback
}