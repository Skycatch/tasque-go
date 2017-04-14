FROM skycatch/tasque:6.10-alpine-onbuild

COPY ./tasque /usr/bin
RUN chmod 777 /usr/bin/tasque
ENTRYPOINT [ "tasque" ]
