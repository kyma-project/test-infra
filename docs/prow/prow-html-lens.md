# HTML Lens

Spyglass HTML lens allows to render HTML files in the job results.

## Usage

All files named `report*.html` saved in the artifacts directory will be rendered on the job results page. See the example:

```bash
golangci-lint  run ./... --out-format html > "${ARTIFACTS}/report-golint.html"
```
