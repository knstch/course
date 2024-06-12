FROM postgres:16.2

COPY postgresql.conf /etc/postgresql.conf

EXPOSE 5432/tcp

## Прописываем отельно путь к конфигу, т. к. по дефолту он расположен вместе с папкой БД /var/lib/postgresql/data/, 
## которая монтируется как том и не позволяет применить новый конфиг.
CMD ["postgres", "-c", "config_file=/etc/postgresql.conf"]

