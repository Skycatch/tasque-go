FROM tasque/node:6.10-alpine

COPY ./tasque /usr/bin
RUN chmod 777 /usr/bin/tasque
ENTRYPOINT [ "tasque" ]
