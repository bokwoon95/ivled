# ivled
[![asciicast](https://asciinema.org/a/EwQph5N9EHTifKmH4CpdSQYdj.png)](https://asciinema.org/a/EwQph5N9EHTifKmH4CpdSQYdj)
IVLE downloader lite. Simple command line wrapper over IVLE LAPI. There is no package installation needed, everything is in one binary file. There is no diving into source code to set some hardcoded API keys/Auth tokens, ivled sets up its own config file together with you. It's fast and it's lightweight✌️.

Tested on macOS Mojave, <s>Windows support coming</s> Windows support delayed because of some weird issues and I really don't feel like developing on Windows right now.

You can exclude folders by name, and files by filetype or name. Just run `ivled config` to open your config file. The config file is hopefully self explanatory enough, otherwise just [drop me an issue](https://github.com/bokwoon95/ivled/issues). By default, [mp4 mp3 mov avi] files are excluded.

# How to use
`ivled`        : Downloads your latest IVLE files into a directory based on your config file.
               If your config file is absent, it will run you through the configuration process
               
`ivled config` : Opens your config file with an external text editor.
               If your config file is absent, it will run you through the configuration process
               
`ivled reset`  : Deletes your config file

`ivled help`   : Displays this help

# macOS Installation
1. Use Chrome (not Safari) to download the [ivled file](https://github.com/bokwoon95/ivled/blob/master/ivled). Make sure it downloads into your 'Downloads' folder.
2. Open Terminal. Run this command
```
mv ~/Downloads/ivled /usr/local/bin
chmod a+x /usr/local/bin/ivled
```
3. Run `ivled` to begin.

If you want to inspect or edit your config anytime, just run the command `ivled config`. <sub>(JSON is not a friendly format to edit config files in, I plan to eventually move over to TOML.)</sub>
