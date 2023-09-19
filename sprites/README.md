# Create sprites

1. Change the source file (e.g. the SVG file)
2. Create a PNG file from the source file
    * SVG to PNG: `inkscape --export-filename=dash-pattern.png dash-pattern.svg`
3. Merge the new PNG into the `sprite.png` via the GIMP file `sprite.xcf` and adjust `sprite.json` accordingly (only for new sprites or changes sizes/locations)
