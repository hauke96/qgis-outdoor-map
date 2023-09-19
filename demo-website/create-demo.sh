#!/bin/bash

OUTPUT="./dist"

echo "Ensure output folder"
rm -rf $OUTPUT
mkdir $OUTPUT

echo "Copy tiles"
cp -r ../tiles $OUTPUT

echo "Copy style"
cp ../style.json $OUTPUT/

echo "Adjust URL in style"
sed -i 's/http:\/\/localhost:8000/https:\/\/hauke-stieler.de\/public\/maps\/osm-outdoor-map-demo/g' $OUTPUT/style.json

echo "Copy index.html"
cp index.html $OUTPUT/

echo "Copy sprites"
mkdir $OUTPUT/sprites
cp ../sprites/sprite.json $OUTPUT/sprites
cp ../sprites/sprite.png $OUTPUT/sprites

echo "Demo website prepared"
