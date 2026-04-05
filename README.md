# Hey Listen

Hey Listen is a Terminal User Interface (TUI) client for YouTube Music written in Go. It allows you to browse your favorite playlists, albums, and tracks, and listen to music directly from your terminal with a minimalist and efficient interface.

## Features

- List liked playlists, albums, and tracks.
- Track playback with a progress bar and volume control.
- Automatic caching system for songs.
- Prefetching (early download) of upcoming tracks in the queue.
- Modular and consistent interface.

## Requirements

Before you begin, ensure you have the following installed:

1. Go (version 1.25.7 or superior).
2. yt-dlp: To download audio streams.
3. FFmpeg: Required for audio extraction and conversion by yt-dlp.
4. Audio Libraries: On Linux, you may need ALSA or PulseAudio development headers (e.g., libasound2-dev on Debian/Ubuntu).
5. Nerd Font: Required to correctly render icons used in the interface (e.g., JetBrainsMono Nerd Font).

## Recommendations

For the best experience, it is highly recommended to use a modern, GPU-accelerated terminal emulator. This type of terminal provide better performance and correct rendering of the TUI components.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/maiconpml/heylisten.git
   cd heylisten
   ```

2. Build the project:

   ```bash
   go build -o heylisten main.go
   ```

## Configuration and Authentication

To access your private playlists and liked music, Hey Listen requires your YouTube Music authentication cookie.

1. Open YouTube Music (music.youtube.com) in your browser.
2. Open Developer Tools (F12).
3. Go to the Network tab and refresh the page.
4. Search for any request to music.youtube.com and copy the value of the cookie field in the Headers.
5. Run the program for the first time:

   ```bash
   ./heylisten
   ```

6. Paste the cookie string when prompted. It will be saved to ~/.config/yt-music-tui/cookie.txt.

## Usage

After the initial configuration, simply run:

```bash
./heylisten
```

### Keyboard Shortcuts

#### Navigation

- 1, 2, 3: Go to Home, Library, or Search tabs (Search is currently under development).
- Tab / Shift+Tab: Next/Previous tab.
- Enter: Select playlist, music, or album.
- Esc / Backspace: Go back.
- ] / [: Next/Previous section.
- k / j or Up / Down: Move selection up/down.

#### Playback

- Space: Play/Pause.
- p: Play selected item.
- P: Play selected item randomly (shuffle).
- n / N: Next/Previous track.
- - / -: Volume up/down.

#### General

- ?: Show/Hide keybindings help.
- q / Ctrl+C: Quit.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
