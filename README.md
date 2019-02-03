# ivled
IVLE downloader lite. Simple command line wrapper over IVLE LAPI.
All the source code is in a single `ivled.go` file because the IVLE API is not complicated. You just need to make some HTTP requests to obtain a file list, parse some JSON, make some more HTTP requests to download the files. (You don't need OOP 🙄)

Tested on macOS Mojave, Windows support coming. By default, [mp4 mp3 mov avi] files are excluded, but support for configuring ignorable filetypes (and ignorable folders) is also planned.

# macOS Installation
1. Use Chrome (Safari downloads .dms files instead) to download ivled file (marked by "<-- macOS USERS DOWNLOAD THIS"). Make sure it downloads into your 'Downloads' folder.
2. Open Terminal. Run this command
```
mv ~/Downloads/ivled /usr/local/bin
chmod a+x /usr/local/bin/ivled
```
3. That's it. Run `ivled` to begin.
