# ivled
[![asciicast](https://asciinema.org/a/EwQph5N9EHTifKmH4CpdSQYdj.png)](https://asciinema.org/a/EwQph5N9EHTifKmH4CpdSQYdj)
IVLE downloader lite. Simple command line wrapper over IVLE LAPI. There is no package setup (looking at you python), everything is in one binary file. There is no diving into source code to set some environment variables, ivled reads from its own config.json which it will set up together with you. It's fast and it's lightweight, and it might sometimes give you segmentation faults✌️.

Tested on macOS Mojave, <s>Windows support coming</s> Windows support delayed because of some weird issues and I really don't feel like developing on Windows right now.

By default, [mp4 mp3 mov avi] files are excluded, but you can also add and remove filetypes from your config file. Just run `ivled config` to begin.  
You can also add folders and files you want to ignore to your config file. Right now there's no explicit documentation on how to do that, but if you know JSON you can figure it out. If you don't know JSON just hang tight.

# How to use
`ivled`        : Downloads your latest IVLE files into a directory based on your config file.
               If your config file is absent, it will run you through the configuration process
               
`ivled config` : Opens your config file with an external text editor.
               If your config file is absent, it will run you through the configuration process
               
`ivled reset`  : Deletes your config file

`ivled help`   : Displays this help

# macOS Installation
1. Use Chrome (not Safari) to download ['ivled' file](https://github.com/bokwoon95/ivled/blob/master/ivled). Make sure it downloads into your 'Downloads' folder.
2. Open Terminal. Run this command
```
mv ~/Downloads/ivled /usr/local/bin
chmod a+x /usr/local/bin/ivled
```
3. Run `ivled` to begin.

If you want to inspect or edit your config anytime, just run the command `ivled config`. <sub>(JSON is not a friendly format to edit config files in, I plan to eventually move over to TOML.)</sub>
