# Sana

A simple terminal-based expense tracker built with Go.

![Sana Screenshot](doc/sana.png)

## Features

- Track expenses with date, description, category, and amount
- View expense summary by category
- Terminal user interface (TUI)

## Usage

### Run the development application

```bash
make run
```

### Run the production application

```bash
make run-prod
```

or

```bash
make install
sana
```

## Keybindings

### Add box

- `tab` - Autocomplete suggestion
- `shift+tab` - Move to previous field
- `down` - Move to next field
- `up` - Move to previous field
- `enter` - Submit form
- `esc` - Cancel form

### Global keybindings

- `a` - Select Add box
- `s` - Select Summary box
- `e` - Select Expenses box
- `j` - Move selection down
- `k` - Move selection up
- `down` - Move selection down
- `up` - Move selection up
- `g` - Move selection to top
- `G` - Move selection to bottom
- `home` - Move selection to top
- `end` - Move selection to bottom
- `r` - Refresh data
- `q` / `ctrl+c` - Quit
