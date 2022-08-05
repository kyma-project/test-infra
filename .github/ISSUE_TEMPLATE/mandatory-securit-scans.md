---
name: Mandatory Securit Scans
about: Scan newly created repository.
title: ''
labels: area/security
assignees: ''

---

Please follow detailed process description: https://wiki.one.int.sap/wiki/display/kyma/Add+new+repositories+to+the+mandatory+scanners

Tasks:
-[ ] Add Whitesource Prow Job. [requestor]
-[ ] Protecode: provide a list of images to scan. [requestor]
-[ ] Protecode: update the Jenkins pipeline. [@neighbors]
-[ ] Checkmarx: create projects for each new GitHub Repository. [@neighbors]
-[ ] Checkmarx: add new project ids to Kyma Security Scan Jenkins pipeline. [@neighbors]
