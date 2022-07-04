#!/usr/bin/env bash



# golangci-lint  run ./... --out-format html > "${ARTIFACTS}/report-golint.html"

cat <<EOF > "${ARTIFACTS}/report-example.html"
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>HTML title</title>
        <style>
            body {background-color: white;}
        </style>
    </head>
    <body>
        <h1>Example report</h1>
        <p>This is an example html file to show that more than one HTML file can be shown at the same time.</p>
    </body>
</html>
EOF
