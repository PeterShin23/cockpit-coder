module github.com/PeterShin23/cockpit-coder/backend

go 1.21

require (
	github.com/creack/pty v1.8.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/mux v1.8.1
	github.com/PeterShin23/cockpit-coder/backend/internal/auth v0.0.0-00010101000000-000000000000
	github.com/PeterShin23/cockpit-coder/backend/internal/session v0.0.0-00010101000000-000000000000
	github.com/PeterShin23/cockpit-coder/backend/internal/pty v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/kr/pty v1.7.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.21.0 // indirect
)
