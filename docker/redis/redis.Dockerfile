FROM redis:latest

ENV REDIS_PASSWORD password

EXPOSE 6379

COPY redis.conf /usr/local/etc/redis/redis.conf

CMD ["redis-server", "/usr/local/etc/redis/redis.conf", "--requirepass", "${REDIS_PASSWORD}"]