# ivled
IVLE downloader lite. Simple command line wrapper over IVLE LAPI.
All the source code is in a single `ivled.go` file because the IVLE API is not complicated. You just need to make some HTTP requests to obtain a file list, parse some JSON, make some more HTTP requests to download the files. (You don't need OOP 🙄)

Tested on macOS Mojave, Windows support coming. By default, [mp4 mp3 mov avi] files are excluded, but support for configuring ignorable filetypes (and ignorable folders) is also planned.

# macOS Installation
1. Use Chrome (not Safari) to download 'ivled' file (marked by "<-- macOS USERS DOWNLOAD THIS"). Make sure it downloads into your 'Downloads' folder.
2. Open Terminal. Run this command
```
mv ~/Downloads/ivled /usr/local/bin
chmod a+x /usr/local/bin/ivled
```
3. Run `ivled` to begin. You will go through an initial setup process, after that your config settings will be written to `~/.config/ivled.json`. The next time, just run `ivled` to pull the latest files from IVLE.

If you want to inspect or edit your config anytime, just run the command `open ~/.config/ivled.json`. <sub>(JSON is not a friendly format to edit config files in, I plan to eventually move over to TOML.)</sub>
