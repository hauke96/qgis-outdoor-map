## Prerequisites

* Make sure Chromium is installed and the `chromium` command is available from the command line
* Make sure the `pdfcropmargins` command is available from the command line
  * Arch Linux: Install the AUR `pdfcropmargins` package
  * Others: It's a python package, use your package manager or pip to install it

## Generate legend graphic

1. Make sure the TODO file contains the correct data you want to have in your legend
2. Call the Script: `./generate-legend.sh TODO-FILE`

## Useful commands

### Get all style entries as sorted list

```shell
cat ../style.json | grep "\"id\"" | sed 's/.*"id": "//g' | sed 's/",.*$//g'
```