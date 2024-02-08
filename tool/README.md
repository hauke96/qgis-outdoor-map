This golang project implements some additional tools that are needed for a good-looking outdoor map.

# Preprocessor

This tool does the following things but does _not_ filter the data.
It just creates new things and changes tags to make styling easier.

* Handling nodes
  * All nodes are kept and stay unchanged
* Handling ways
  * For some types of ways (e.g. ford and shops), nodes at their centroid are created.
    This makes it easier to create uniform POI styling because in OSM some of these things are already nodes and some are modeled as ways.
  * Highway tags are adjusted for simplicity. All ways that are not accessible (e.g. due to constructions) get `access=no`.
  * All ways being part of any hiking route are tagged accordingly and also receive special hiking-rout-name-tags.
* Handling relations
  * Ways of hiking routes are collected for tagging as described above

# Tile proxy

Because `tileserver-gl` (at least version 4.7.0) is unable to render WebP-based raster tiles for hillshading, the `tile-proxy` command starts a proxy server that is able to convert WebP images into PNG images, which are then usable by tileserver-gl. 

# TODOs

(currently no TODOs are known for the tool)