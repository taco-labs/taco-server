![AWS Code Build Status](https://codebuild.ap-northeast-2.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoidVAyN0FKR0lUejhpZlVaekpwNzJzR3d1c0F5VllrSzVMclBFQXJPYWx6MlZmR1NpM2JzcGN2dWNUYWpWN01rVFZtZ2ZKNVY3Z1c0Qmk1eTN6dVB2RW8wPSIsIml2UGFyYW1ldGVyU3BlYyI6IlJ3c0NkOGcxSVRwRkFhZWMiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=main)

# Prerequisite
- [pre-commit](https://pre-commit.com/)
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
- [golangci-lint](https://github.com/golangci/golangci-lint)

# TACO backend
Backend impl for backend

## Quick Start
```sh
make run_local

# User API
curl localhost:18881/healthz

# Driver API
curl localhost:18882/healthz

```

## Directory Structure

