# HTML lens

Spyglass HTML lens allows developer to display HTML files in the job result

## Usage

All files named `report*.html` saved in the artifacts directory by the prowjob will be displayed on the job results page. See the example:

```bash
golangci-lint  run ./... --out-format html > "${ARTIFACTS}/report-golint.html"
```
