# ivled
IVLE downloader lite. Simple command line wrapper over IVLE LAPI.
All the source code is in a single `ivled.go` file because the IVLE API is not complicated. You just need to make some HTTP requests to obtain a file list, parse some JSON, make some more HTTP requests to download the files. (You don't need OOP 🙄)

Tested on macOS Mojave, Windows support coming. By default, [mp4 mp3 mov avi] files are excluded, but support for configuring ignorable filetypes (and ignorable folders) is also planned.
