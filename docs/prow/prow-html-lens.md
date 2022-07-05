# HTML lens

Spyglass HTML lens allows to display HTML files in the job results.

## Usage

All files named `report*.html` saved in the artifacts directory will be displayed on the job results page. See the example:

```bash
golangci-lint  run ./... --out-format html > "${ARTIFACTS}/report-golint.html"
```
