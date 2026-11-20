package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"yerss/internal"
	"yerss/internal/config"
	"yerss/internal/feed"
)

// Exit codes for the startup gate refusals. Existing load/DB errors remain
// exit 1; these distinct codes let scripts and tests tell the two refusal
// reasons apart.
const (
	exitEmptyFeeds = 2
	exitNoWorking  = 3
	verifyPerFeed  = 10 * time.Second
	verifyOverall  = 30 * time.Second
)

// runStartupGate verifies that at least one RSS feed is configured and live
// before the TUI takes over the terminal. It prints the reason for any
// refusal to stderr and returns the process exit code, or 0 to proceed.
//
//   - empty/missing feeds file            -> exit 2, naming the file
//   - feeds table already has a verified feed -> 0 (offline-safe, no network)
//   - first run: fetch with 10s/feed, 30s overall; no feed parsed -> exit 3
func runStartupGate(cfg *config.Config, st *internal.Store) int {
	urls, err := feed.LoadFeeds(cfg.FeedsFile())
	if err != nil {
		// A missing file is the "nothing configured yet" case (exit 2). Any
		// other read failure is distinct: report it clearly and exit 1.
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "yerss: no RSS feeds configured; add at least one feed URL to %s\n", cfg.FeedsFile())
			return exitEmptyFeeds
		}
		fmt.Fprintf(os.Stderr, "yerss: cannot read feeds file %s: %v\n", cfg.FeedsFile(), err)
		return 1
	}
	if len(urls) == 0 {
		fmt.Fprintf(os.Stderr, "yerss: no RSS feeds configured; add at least one feed URL to %s\n", cfg.FeedsFile())
		return exitEmptyFeeds
	}

	verified, err := st.HasVerifiedFeed()
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot check verified feeds: %v\n", err)
		return 1
	}
	if verified {
		return 0
	}

	fmt.Fprintln(os.Stderr, "verifying feeds...")
	res, err := feed.VerifyFeedsWithProgress(st, urls, verifyPerFeed, verifyOverall, func(outcome feed.FeedOutcome) {
		if outcome.Err != nil {
			fmt.Fprintf(os.Stderr, "  failed %s: %v\n", outcome.URL, outcome.Err)
			return
		}
		fmt.Fprintf(os.Stderr, "  ok %s (%d articles)\n", outcome.URL, outcome.Articles)
	})
	if err != nil || res.Feeds == 0 {
		fmt.Fprintf(os.Stderr, "yerss: no working RSS feed found; check the URLs in %s\n", cfg.FeedsFile())
		return exitNoWorking
	}
	fmt.Fprintf(os.Stderr, "verified %d of %d feeds\n", res.Feeds, len(urls))
	return 0
}
