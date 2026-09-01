## 1. Session type

- [x] 1.1 Create `internal/app` with a `Session` holding store, config, and image caches
- [x] 1.2 Add delegating methods for the store calls the model uses today

## 2. Wire the model

- [x] 2.1 Change `ui.New` to accept `*app.Session` and store it on `Model`
- [x] 2.2 Route the model's store/config/image access through the session
- [x] 2.3 Update `cmd/yerss/main.go` to construct and pass the session

## 3. Validation

- [x] 3.1 Run `go test ./...` and `go build ./...`
