## 1. Palette role

- [x] 1.1 Add `Selection` to `Palette` and set it in both palettes
- [x] 1.2 Replace the hardcoded `#707070` selection backgrounds with the role

## 2. Precomputed styles

- [x] 2.1 Add precomputed style fields to the model and populate them in `New`
- [x] 2.2 Swap per-frame style construction for the cached styles

## 3. Validation

- [x] 3.1 Run `go test ./internal/ui/...` (render tests must stay green with byte-identical output)
