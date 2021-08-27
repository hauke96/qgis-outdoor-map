#!/bin/bash

set -e
set -x

export PGPASSFILE=./.pgpass
export PGHOST=localhost
export PGUSER=postgres

psql -c "CREATE EXTENSION hstore;" 2> /dev/stderr || echo "Skip hstore extension creation."

osm2pgsql --create --slim -G --hstore --number-processes 4 "$1"
