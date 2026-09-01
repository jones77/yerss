## 1. Store aggregation

- [x] 1.1 Add an `ImageStats`/`ImageSize` type and a `Store.ImageStats()` method that reads `article_images` and `articles`
- [x] 1.2 Unit-test `ImageStats` on a seeded in-memory database

## 2. CLI flag and report

- [x] 2.1 Register an `--image-stats` flag and wire an early exit path in `main` that runs before the startup gate
- [x] 2.2 Print the report (DB+WAL size, counts, photo-byte min/median/p90/max, block bytes, decoded-size estimate) to stdout
- [x] 2.3 Probe each photo's dimensions via `image.DecodeConfig` and fold the estimate into the report

## 3. Validation

- [x] 3.1 Run `go test ./cmd/yerss/... ./internal/store/...` and `openspec validate measure-image-sizes`
