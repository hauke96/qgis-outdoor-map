#!/bin/python

from qgis.core import *
import sys
import argparse

# Surpress weird error message when creating PNGs
from osgeo import gdal
gdal.PushErrorHandler('CPLQuietErrorHandler')

# Parse CLI arguments
parser = argparse.ArgumentParser(description="Render QGIS layout to a PDF or image file.")
parser.add_argument("-t", "--type", help="The file type to generate. Default: pdf", choices=["pdf", "png"], default="pdf")
parser.add_argument("-o", "--output", help="The output file name. Default: <layout_name>.<type>")
parser.add_argument("-p", "--project", help="The project file to load. Default: ./map.qgs", default="./map.qgs")
parser.add_argument("-d", "--dpi", help="The DPI that should be used. Default: 300", type=int, default=300)
parser.add_argument("layout_name", help="The name of the layout that should be rendered.")
args = parser.parse_args()

if args.output == None:
	args.output = "./" + args.layout_name + "." + args.type

print(f"File type   : {args.type}")
print(f"Layout      : {args.layout_name}")
print(f"DPI         : {args.dpi}")
print(f"Output file : {args.output}")
print()

# Start QGIS application without GUI
qgs = QgsApplication([], False)
qgs.initQgis();

# Load project
project = QgsProject.instance()
project.read(args.project)

# Get the according layout
manager = project.layoutManager()
layout = manager.layoutByName(args.layout_name)

if layout == None:
	print("The layout '" + args.layout_name + "' could not be found.", file = sys.stderr)
	quit()

# Generate the PDF/PNG/... file
exporter = QgsLayoutExporter(layout)
if args.type == 'pdf':
	settings = QgsLayoutExporter.PdfExportSettings()
	settings.dpi = args.dpi
	exporter.exportToPdf(args.output, settings)
elif args.type == 'png':
	settings = QgsLayoutExporter.ImageExportSettings()
	settings.dpi = args.dpi
	exporter.exportToImage(args.output, settings)

# Gracefully close QGIS
qgs.exitQgis();
