# Git hools

## Overview

This directory provides useful git hooks that can be used in standard git development flow.

## Installation

To use those hooks in your have to configure the default git hooks path for your repository.
Run the command below in the root of your `test-infra` git repository:
```shell
git config core.hooksPath .githooks
```
After that your local git repository will use git hooks from `.githooks` path.

## Troubleshooting

### Mac: VCS integration in GoLand can't use applications installed by Homebrew

This is an issue that is known for all JetBrains IDEs which run on a Mac.
Upon startup GoLand will initialize the `PATH` based on the macOS' `launchd` path.
The [solution](https://apple.stackexchange.com/questions/51677/how-to-set-path-for-finder-launched-applications) is to add `/usr/local/bin` path to the system `PATH` variable by using the command listed below.
**This change will be applied to ALL uses of a computer!**
```shell
sudo launchctl config user path /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
```

After running this command reboot your machine to apply changes.

## Hooks list
|Name|Type|Description|
|---|---|---|
|01-rendertemplates.sh|pre-commit|Automatically renders templates and adds the resulted files to the commit if there was any changes.| 
