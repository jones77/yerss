## 1. Documentation

- [x] 1.1 Add a "Dates and times" note to `README.md` stating that publication dates and times (day headers, article rows, and the article top border) are displayed in the user's local timezone, and that the app follows the system/terminal local time (`TZ` environment) with no in-app timezone setting

## 2. Verification

- [x] 2.1 Run `openspec validate --all --strict` and confirm the change validates with `skip_specs: true` (no spec files expected)
