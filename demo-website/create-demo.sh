#!/bin/bash

OUTPUT="./dist"

rm -rf $OUTPUT
mkdir $OUTPUT

cp -r ../tiles $OUTPUT

cp ../style.json $OUTPUT/
sed -i 's/http:\/\/localhost:8000/https:\/\/hauke-stieler.de\/public\/maps\/osm-outdoor-map-demo/g' $OUTPUT/style.json

cp index.html $OUTPUT/

echo "Demo website prepared"
