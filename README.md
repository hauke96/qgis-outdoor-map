# Outdoor map style

This repo contains a Mapbox style file for an OSM-based outdoor map focussed on hiking and trekking.
PBF vector tiles are used (generated with Tilemaker) and styled via Maputnik.

## Online example

Take a look at https://hauke-stieler.de/public/maps/outdoor-hiking-map-demo for a live-demo of a (more or less) up-to-date state of the style.

## Download data

Just download/prepare any `.osm.pbf` file yourself or use the following script for certain pre-defined areas.

Script for pre-defined areas:

1. Go into `data` folder
2. Execute `import-data.sh` script with one of the given area names. They are all listed at the bottom of the script.

## Create vector tiles from PBF file

Just call the `make-tiles.sh data.osm.pbf` script.
It calls the preprocessor to prepare the data, removes the `./tiles` folder, recreates it and fills it with XYZ vector tiles in PBF format.

## Serve tiles locally

The `serve.sh` is used to host a) the tiles, sprites, etc. and b) the Maputnik application.

Tiles need to be served from `http://localhost:8000/tiles` and sprites from `.../sprites` in order to be used with Maputnik.
Use the `serve.sh` script to start the tile server and Maputnik.
No parameters are needed, since the script uses the `./style.json` with the tile and sprite URLs defined in there.

The Maputnik desktop tool is started as well and will automatically save everything to the `style.json` file.

## Edit style with Maputnik

Take a look at Maputnik and Tilemaker if you're not familiar with those tools.

### Maputnik shows old tiles

**Problem:** Maputnik doesn't use the newly generated tiles, data is missing or something else is not up-to-date.

**Solution:** Use a single(!) private-browsing window (incognito-mode or whatever your browser calls it) and open Maputnik there to prevent the browser from using any old caches. Whenever the tiles change, close that window and restart Maputnik in a new private-browsing window.

## Style guide

* **Hiking infrastructure has a higher precedence over non-hiking infrastructure.** Example: Drinking water POIs are already visible at zoom level 12, advanced trails have a bright yellow background and generally all hiking trails are directly recognizable.
* **Hiking relevant data only.** Things, that are not related or important for hikers (or other outdoor enthusiasts) are irrelevant. This includes for example parking spaces. Yes, people arrive by car but why should a map for hiking include parking if you use your phone for car navigation and finding a parking spot?
* **Use as few different colors amd font-styles as possible.** Sometimes, adjustments of font or icon colors are needed for better visualization, but keep that to a minimum.
* **Orientate yourself by the [2014 material design color palette](https://material.io/design/color/the-color-system.html#tools-for-picking-colors).** These colors work quite well but I sometimes change some parameters of the color where appropriate. But if such a color works, then take it.
* **Sprites should be of high quality.** When possible, create large sprites and then use the scale factor to scale them down for the map. This ensures sharp icons and the possibility for hires maps.

_More guidelines will be added over time._ 

## Generate legend graphic

The legend graphic is a PDF file generated using a JSON schema file.
See [`tool` folder documentation](./tool/README.md) for further documentation.

## Style usage in QGIS

The `qgis-map.qgz` file contains layer for the locally hosted vector tiles as well as hillshading and contour lines.

**Pro-tip:** It can also be used to (a bit poorly) create Carto-OSM map with hillshading and contour lines ;)

### Updating QGIS style

Note that the style must be updated manually:

1. Select "localhost" layer
2. Open style settings
3. Go to "Style manager"
4. Click on "Load style" (folder icon)
5. Select the `style.json` file
6. Done

### Known problems with QGIS rendering

* Offsets are reversed. Example: Protected areas have a shading which is done by an offset to the inside of the polygon but in QGIS it's towards the outside.
* Repeating icons aren't always working and QGIS sometimes just places the icon on the center of the line.

# TODOs

* Bus stations
* Label style of lakes based on their size
  * large lake: large font, wide letter spaces, bold
  * mid lake: mid sized font, slight letter space, bold
  * small lake: regular font, no letter space, regular font
* Add motorways and trunks to map
* Rename git-repo since it's not a QGIS-based map anymore
* Evaluate usefulness of `render-layout.py` and `render-all.sh` scripts.
* Tutorial on adding/editing data and layers
* Tutorial on creating sprites
* Update example screenshots and add photo of printed map
* Determine workflow on how to create good-looking print layouts. Maybe create a template layout or something.

---

# DEPRECATED DOCUMENTATION

## Render map

Make sure the project is loaded in QGIS and that you have the data loaded into your database.

1. Create a layout or change an existing one
2. Add the OSM attribution if you use OSM data
3. Use the normal QGIS mechanisms to export the layout to PDF, PNG, ...

### Render via Terminal/CLI

1. Make sure you created your layout in QGIS
2. Use the `render-layout.py` script to render a single layout (ude `render-layout.py --help` for more information)

### Render all pre-defined layouts

There are some pre-defined layouts (such as `a2-fischbeker-heide`) wich legend, scale bar, attribution and everything.
Calling the script `render-all.sh` renders them all as PDF and PNG files, no parameter needed.
