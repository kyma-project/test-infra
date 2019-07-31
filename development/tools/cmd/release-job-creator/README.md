Input Variables:
- OLD_RELEASES: `1.1,1.2` `mandatory`
- NEW_RELEASES: `1.4,1.5` `mandatory`
- REF_RELEASE - `1.3` `mandatory`



Algorithm:
- Traverse prow/jobs/kyma and read all yamls
- Read presubsmits block
- Take out unsupported release extract
- Traverse through jobs definition folders
- Create an old extract by replacing the old name 
- Create a new extract for a release name from supported release
- Remove the old extract
- Add the new extract
- Overwrite the file by adding new jobs(based on the supported flag) and removing old jobs(based on the unsupported flag)

Usage:

```
    OLD_RELEASES=1.1 NEW_RELEASES=1.4 REF_RELEASE=1.3 go run main.go
```