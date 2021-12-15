#!/bin/bash


LAYOUTS=(
	"a2-fischbeker-heide"
	"a2-sachsenwald"
)
OUTPUT="./rendered-maps"

mkdir -p "$OUTPUT"

for l in "${LAYOUTS[@]}"
do
	echo "Render layout '$l' to PDF"
	./render-layout.py -o "$OUTPUT/$l.pdf" "$l"

	echo
	echo "Render layout '$l' to PNG"
	./render-layout.py -o "$OUTPUT/$l.png" -t png "$l"
done

echo
echo "Done."
