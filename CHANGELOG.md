## [1.0.2] - 2026-05-24

### 🐛 Bug Fixes
- *(batch)* Reuse Browser instance and fix goroutine leak on error
- *(ci,batch)* Pin goreleaser to v2 and fix continue-on-error semantics

## [1.0.1] - 2026-05-24

### 🐛 Bug Fixes
- *(snap)* Fix full-page screenshot always capturing viewport size
- *(snap)* Always force color scheme to prevent system theme bleeding into screenshots
- Resolve four correctness bugs across batch, browser, and snap
- *(release)* Handle missing CHANGELOG.md on first release

## [1.0.0] - 2026-05-19

### 🚀 Features
- *(snap)* Define Option types, Format/Device constants and device presets
- *(snap)* Implement Chrome auto-detection with platform-specific fallbacks
- *(snap)* Implement core Capture API with full-page, selector, clip, wait, JS/CSS injection and multi-format output
- *(cli)* Bootstrap root command with global flags
- *(cli)* Add snap command with all capture flags
- *(cli)* Add batch command with concurrency and file pattern support
- *(cli)* Add version command and shared JSON encoder helper
- *(snap)* Add Browser reuse API with CaptureAll and concurrency support


### 🐛 Bug Fixes
- *(snap)* Implement real network idle detection using CDP network events


### 📚 Documentation
- Add README

