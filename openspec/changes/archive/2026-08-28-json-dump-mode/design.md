# Design: json-dump-mode

## Flag + early exit

`cmd/yerss/main.go` adds the flag and short-circuits before the startup gate and
TUI, but after the `-c`/`-e` edit flow and after `store.Open`:

```go
var jsonOut bool
pflag.BoolVarP(&jsonOut, "json", "j", false, "dump articles as JSON and exit")
...
if code, done := runEditFlags(editConfig, editFeeds); done { os.Exit(code) }
cfg, err := config.Load("")
st, err := store.Open(cfg.DBPath())
...
if jsonOut {
    os.Exit(runJSONDump(cfg, st))
}
if code := runStartupGate(cfg, st); code != 0 { os.Exit(code) }
```

`-j` does not early-exit at parse time (unlike `-c`/`-e`), because it needs the
loaded config and store. It runs after the edit flags so `-e`/`-c` edits apply.

## Grouping

`store.ListArticles("")` already returns articles reverse chronological
(`ORDER BY COALESCE(published_at,0) DESC, id DESC`), so a single pass buckets
into date-ordered groups with articles already in the right order:

```go
type articleJSON struct {
    ID          int64    `json:"id"`
    FeedURL     string   `json:"feed_url"`
    GUID        string   `json:"guid"`
    Title       string   `json:"title"`
    Link        string   `json:"link"`
    Author      string   `json:"author"`
    PublishedAt string   `json:"published_at"`
    Content     string   `json:"content"`
    Description string   `json:"description"`
    Read        bool     `json:"read"`
    FetchedAt   string   `json:"fetched_at"`
    Categories  []string `json:"categories"`
}

type dateBucket struct {
    date string         // "YYYYMMDD", local
    arts []articleJSON
}
```

```go
arts, _ := st.ListArticles("")
var buckets []dateBucket
idx := map[string]int{}
for _, a := range arts {
    t := a.PublishedAt
    if t.IsZero() { t = a.FetchedAt }      // undated fallback
    key := t.Local().Format("20060102")
    i, ok := idx[key]
    if !ok {
        i = len(buckets)
        buckets = append(buckets, dateBucket{date: key})
        idx[key] = i
    }
    buckets[i].arts = append(buckets[i].arts, toArticleJSON(a))
}
```

Because `ListArticles` is desc, the first time a date key is seen is its
most-recent occurrence, so `buckets` ends up in descending date order and each
bucket's articles are already descending. No re-sort needed.

## Ordered object output

Go's `encoding/json` emits `map[string]…` keys in ascending lexicographic order,
which is chronological — the opposite of what we want. So `dateBuckets` gets a
custom marshaler that writes keys in slice (descending) order:

```go
func (b dateBuckets) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte('{')
    for i, g := range b {
        if i > 0 { buf.WriteByte(',') }
        kb, _ := json.Marshal(g.date)
        buf.Write(kb)
        buf.WriteByte(':')
        ab, _ := json.Marshal(g.arts)
        buf.Write(ab)
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}
```

Empty store → `buckets` is nil → the marshaler writes `{}`. Exit 0.

## Field formatting

`published_at` and `fetched_at` are emitted as RFC3339 strings
(`t.Format(time.RFC3339)`, preserving the offset) rather than raw Unix integers,
so they are self-describing. `categories` is the article's `[]string` (empty
array `[]` when none, not `null`, by initializing the slice).

## Edge cases

- **Undated articles** (`published_at` zero): fall back to `fetched_at` for the
  date key so they still appear under a real day rather than `19700101`.
- **Combined `-j -e`:** `runEditFlags` edits feeds first, then the dump runs
  against the existing DB (the dump does not refresh, per spec).
- **Combined `-j -c`:** config is edited first; the dump uses the reloaded
  config (e.g. a custom DB path).

## Tests

- `runJSONDump` on a store with articles on three days → JSON object with three
  keys in descending order; each array descending.
- All stored fields present and correctly typed; `categories` is `[]` when empty.
- Local timezone keying: an article at midnight UTC keys to the local day.
- Empty store → `{}` and exit 0.
- `published_at`/`fetched_at` are RFC3339 strings.
