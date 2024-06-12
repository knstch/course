FROM postgres:16.2

COPY postgresql.conf /etc/postgresql.conf

EXPOSE 5432/tcp

CMD ["postgres", "-c", "config_file=/etc/postgresql.conf"]

