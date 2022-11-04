How to create nice looking hillshade tiles using QGIS and GDAL.

# Prerequisites

1. Make sure QGIS is installed and works
2. Make sure GDAL is installed and the `gdal_translate` and `gdal_merge` commands are available

# Download SRTM data

1. Open QGIS
2. Install the SRTM-Plugin (if not already installed)
3. Navigate the the area you want the data from
4. Open the Plugin (Plugins → SRTM Downloader → SRTM Downloader)
5. "Set canvas extent" (or enter manual extent for the data)
6. Select output path (we do need the files later)
7. "Download"
8. Enter credentials (create an account if you don't have one)
9. Close the window when the download completed

We now have the raw SRTM elevation data.

# Optional: Merge multiple SRTM files

Not needed if you only downloaded one file ;)

You probably downloaded multiple files, so we need to merge them.
You have now two options: Use GDAL from terminal or use QGIS.

## Option 1: Merge GeoTiff files using QGIS

1. Open toolbox
2. Search for "merge" (or choose GDAL → Raster miscellaneous → Merge)
3. Select all SRTM layers with the data you just downloaded
4. "Run" and close window after the merge completed
5. Right click on new "Merged" layer → Export → Save As
6. Make sure "GeoTIFF" is the output format
7. Choose a filename
8. "OK"

## Option 2: Merge using GDAL from terminal

1. Open a terminal and navigate to the folder where the raw SRTM files are
2. Use the following command to merge the files: `gdal_merge.py -o merged.tif <input-files>`

# Clipping

Clip the data to the exact extent (each SRTM file is quite large).
Make sure the merges data is available as one layer.

1. Open toolbox
2. Search for "clipping" or select GDAL → Raster extraction → Clip raster by extent
3. Select layer with the merged data
4. Select extent (entry "Use Current Map Canvas Extent" in drop-down menu)
5. Select an output file (like `clipped.tif`; we need that file later)
6. "Run"

# Smoothing + upscaling

The resolution of SRTM data is not so nice and the raw files might be quite noisy (even though you don't see that now).
For SRTM at least upscaling by a factor of 2 works very well.

1. Open a terminal and navigate to the folder where your clipped SRTM file is
2. Smoothing:
  1. Use `gdal_translate -outsize 50% 50% -r bilinear clipped.tif downscaled.tif` to downscale the image
  2. Use `gdal_translate -outsize 200% 200% -r bilinear downscaled.tif normal-scaled.tif` to scale the image back to the original resolution
3. Upscaling:
  1. Use `gdal_translate -outsize 200% 200% -r bilinear normal-scaled.tif upscaled.tif` to upscale the image

This down- and double upscaling automatically smoothes the image. Use other values than 50% and 200% for more/less smoothing and upscaling.
One single large upscaling step from 50% to 200% resolution would create ugly blocks because the interpolation doesn't work nice then.

# Create hillshading

1. Load the upscaled data into QGIS
2. Open toolbox
3. Search for "hillshade" or select GDAL → Raster analysis → Hillshade
4. Select the layer with the upscaled data
5. Select 0.00001 as "Z factor" (maybe a different value works better for you, just play around with this value)
6. Optional: Select "Multidirectional shading" (I like this parameter, it also puts light on very shaded areas)
7. "Run"
8. Close the window

Done, you now have an already nice looking hillshade layer.

## Optional: Styling the hillshade

If you want to use your hillshade layer to add that hillshading to another map, there are some styling adjustments that makes this hillshade even more beautiful.

1. Go into the styling editor for the hillshade layer (select hillshade layer → F7)
2. Change "Brightness" to 30 (or any other value that looks nice)
3. Change "Contrast" to -15 (or any other value that looks nice)
4. Go into the tab "Transparency"
5. Set transparency to 65% (or any other value that looks nice)

You can change several other parameters until you get a nice decent and not too contrasty layer.

Optional: Try your style changed and add the hillshading to the osm.org Carto map:

1. Add the normal osm.org map (or any other you like) as background map
2. Go into the styling editor for the hillshade layer (select hillshade layer → F7)
3. Change "Blending Mode" to "Multiply"

# Create contour lines

1. Load your upscaled data into QGIS (if not already there)
2. Open the toolbox
3. Search for "contour" or select GDAL → Raster extraction → Contour
4. Select your upscaled layer
5. Select an interval between the lines (e.g. 25 to have a line every 25 meters)
6. Use the attribute name ELEV
7. "Run"

Done, you now have okay-looking contour lines as vector features.

## Optional: Smooting

**Notice:** Smoothing requires a lot of memory, so save your project to not loose anything when your computer gets stuck ;)

1. Open the toolbox
2. Search for "smooth" or select Vector geometry → Smooth
3. Select the number of iterations (more iterations = more RAM usage but also more smoothness)
4. "Run"

# Export

Use the normal QGIS export feature for most formats: Rightlick on the layer → Export → Save As.

## Hillshade

Good format here are either raster tiles like XYZ-tiles (using GDAL from the toolbox) or a GeoTIFF file.

## Contour lines

Good format here is any vector format (GeoJSON, ShapeFile, GeoPackage, PBF, ...) or vector tiles like XYZ-tiles (using GDAL from the toolbox).

**Important:** If you want to use the contour lines in this QGIS map projekt, make sure it's a **GeoPackage** with the layer name **contour**.

# Import in this QGIS outdoor map

1. Hillshading
  1. Export your hillshading into a `.tif` file.
  2. Copy the hillshading file next to the QGIS project file and name it `hillshade.tif`.
2. Contours
  1. Export the contours as `.gpkg` and name the contour layer `contour`.
  2. Copy the file next to the QGIS project file and name it `contours.gpkg`.

Now you can open the QGIS project and you'll have your contours and hillshading.
