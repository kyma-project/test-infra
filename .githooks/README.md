# Git hooks

## Overview

This directory provides Git hooks that you can use in a standard Git development flow.

## Installation

To use Git hooks in your local Git repository, configure the default Git hooks path.
To do so, run this command in the root of your `test-infra` Git repository:
```shell
git config core.hooksPath .githooks
```
Now your local Git repository uses Git hooks from the `.githooks` path.

## Troubleshooting

### Mac: VCS integration in GoLand can't use applications installed by Homebrew

This is an issue that is known for all JetBrains IDEs which run on a Mac.
Upon startup, GoLand will initialize the `PATH` based on the macOS' `launchd` path.
The [solution](https://apple.stackexchange.com/questions/51677/how-to-set-path-for-finder-launched-applications) is to add `/usr/local/bin` path to the system `PATH` variable by using the command listed below.
```shell
sudo launchctl config user path /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
```
>**CAUTION:** This change will be applied to ALL uses of a computer!

After running this command, reboot your machine to apply changes.

## Hooks list
|Name|Type|Description|
|---|---|---|
|01-rendertemplates.sh|pre-commit|Automatically renders templates and adds the resulted files to the commit if there were any changes.| 
