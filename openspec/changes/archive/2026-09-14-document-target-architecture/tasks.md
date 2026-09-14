## 1. Architecture overview

- [x] 1.1 Document the target `internal/ui` layout (Model as wiring layer; `listview`, `popup`, `imageview`, `articleview`; `render`/`compose` absorbing pure helpers) in design.md
- [x] 1.2 Document the target `internal/app.Session` responsibilities (feed orchestration down, read-state policy up) in design.md
- [x] 1.3 Record the dependency-direction rule (views may import render/compose; nothing imports Model) in design.md

## 2. Execution order

- [x] 2.1 Fix the order C → A → B → D → E with rationale (low-risk first, A is the lynchpin, D/E swappable last) in design.md
- [x] 2.2 List the per-step correctness invariants (decode-once, O(N) render bounds, snap guards never bounce, one clear per no-native transition) in design.md

## 3. Build-agent handoff prompt

- [x] 3.1 Write a copy-paste build-agent prompt into design.md that names the five change directories in order and directs `/opsx-apply` per change

Run go build ./...
