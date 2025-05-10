FROM postgres:17.3
ADD database.sql /docker-entrypoint-initdb.d/database.sql
ADD insert.sql /docker-entrypoint-initdb.d/insert.sql
RUN localedef -i fr_FR -c -f UTF-8 -A /usr/share/locale/locale.alias fr_FR.UTF-8
ENV LANG fr_FR.utf8
