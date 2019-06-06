# Vulnerability Collector

## Overview

This command outputs vulerabilites in a given container.

## Usage

To run it, use:

```bash
go run cmd/vulnerablilitycollector/main.go  -url=https://{FULL_PATH_TO_CONTAINER_IMAGE}
```

## Output
WARN[0002] Severity HIGH libtasn1-6 4.13                
WARN[0002] Severity HIGH libxrender 0.9.10    
.........
WARN[0002] Severity HIGH dpkg 1.19.0.5ubuntu2.1         
WARN[0002] Severity HIGH libxrender 0.9.10              
WARN[0002] Number of High issues 16                     
INFO[0002] Number of issues 106     