FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# add nobody user, stolen from busy box examples
RUN echo "nobody:x:1:1:nobody:/:/bin/sh" >> /etc/passwd
RUN echo "nobody:x:1:" >> /etc/group

# copy migration files for migration tool to be useful
COPY db/migrations db/migrations

# copy build output to entrypoint dir
COPY bin/app-nix-64 /bin/{{.AppName}}

USER nobody
ENTRYPOINT [ "{{.AppName}}" ]
