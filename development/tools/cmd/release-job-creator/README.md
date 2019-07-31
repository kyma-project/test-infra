Input Variables:
- pre/all: default: all - optional
- unsupported_releases: default - mandatory, comma separated list
- new_release - mandatory



Algorithm:
- Traverse prow/jobs/kyma and read all yamls
- Read presubsmits block
- Take out unsupported release extract
- Note the folder structure after "kyma/"
- Create a name by appending foldernames separated by dashes.
- Create a new name: "pre" "rel" <supported-release-without-dot> <kyma> <folder1-folder2...> 
- Create an old extract by replacing the old name 
- Create a new extract for a release name from supported release
- Add the new extract
- Remove the old extract
- Overwrite the file by adding new jobs(based on the supported flag) and removing old jobs(based on the unsupported flag)
- Handle control exit