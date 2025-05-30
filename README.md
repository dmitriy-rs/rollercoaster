# rollercoaster :roller_coaster:

Running scripts without need to know the task/script manager. Roll on them like a rollercoaster!

## Installation

### Homebrew

```sh
brew tap dmitriy-rs/tap
brew install rollercoaster
```

### Using Go

```sh
go install https://github.com/dmitriy-rs/rollercoaster@latest
```

### From source

```sh
# Clone repo 
git clone https://github.com/dmitriy-rs/rollercoaster
cd rollercoaster

# Build and install
go build -o rollercoaster ./main.go
```

Then place the binary somewhere in your PATH.
```sh
# For example
sudo mv rollercoaster /usr/local/bin/rollercoaster
```

## Usage

Type the name of the command to see all available actions in the current scope
```sh
rollercoaster
```

Type some letters from the command to run one of the commands
```sh
# it will trigger pnpm lint command
rollercoaster li
# or ever will trigger the same action
rollercoaster l
```

That's so simple as that :) 

### Alias

I suggest to create alias in your shell for the command. Something short and handy, I use `rc`
```zsh
# ~/.zshenv
alias rc="rollercoaster"
```

## TODO

### Pre-release tasks

- [x] Default js workspace tasks (add, remove, install, npx)
- [ ] ~CLI flag --cwd to start cli in different target directory~
- [ ] UI with task selection
- [ ] If multiple task matches the query show the same selection UI with mached tasks
- [ ] --accept-first config to always select first match instead of showing the UI
- [ ] Show which letters where matched in UI
- [x] Fuzzy search on mistakes if `ilt` provided `lint` should be selected if available
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