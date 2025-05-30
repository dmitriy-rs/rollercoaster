# rollercoaster :roller_coaster:

Running scripts without need to know the task/script manager. Roll on them like a rollercoaster!

## Installation

WIP

## TODO

### Pre-release tasks

- [x] Default js workspace tasks (add, remove, install, npx)
- [ ] ~CLI flag --cwd to start cli in different target directory~
- [ ] UI with task selection
- [ ] If multiple task matches the query show the same selection UI with mached tasks
- [ ] --accept-first config to always select first match instead of showing the UI
- [ ] Show which letters where matched in UI
- [ ] Fuzzy search on mistakes if `ilt` provided `lint` should be selected if available
- [ ] Add Bun and Deno support

### Think

- [ ] Fallback js manager
- [ ] Nested lock files with different package managers (why?)

### On-release tasks

- [ ] Setup pipeline for release into different platforms
- [ ] Provide good README
- [ ] Provide good installation guide for different platforms
- [ ] Add coverage, release and platform badges

### After initial release

- [ ] Investigate into shell integration
- [ ] Integrate with zsh shell and test usability
- [ ] Add README for shell usage
- [ ] Global command to enable/disable shell intergation
- [ ] Add support to bash itegration
- [ ] Investigate NX support
- [ ] Investigate Makefile support
- [ ] Infinite bug fixes