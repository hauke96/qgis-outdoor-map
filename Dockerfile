FROM postgis/postgis:12-3.1-alpine

#ENV POSTGRES_USER=postgres
#ENV POSTGRES_PASSWORD=secret

ENV PGHOST=localhost
ENV PGUSER=postgres
ENV PGPASSFILE=/.pgpass

COPY ./.pgpass /.pgpass

#RUN apk install sudo
RUN su -c "pg_ctl start" postgres

#RUN createuser osm
RUN createdb -h localhost -U postgres -E UTF8 -O postgres gis
RUN psql -d gis -c "CREATE EXTENSION postgis;"
RUN psql -d gis -c "CREATE EXTENSION hstore;"
RUN psql -d gis -c "ALTER TABLE geometry_columns OWNER TO osm;"
RUN psql -d gis -c "ALTER TABLE spatial_ref_sys OWNER TO osm;"
RUN psql -c "ALTER USER osm PASSWORD 'secret'"

RUN osm2pgsql -d gis --create --slim -G --hstore --number-processes 4 /data.osm.pbf
