FROM debian:stable-slim

RUN mkdir /usr/local/lib/cockroach /cockroach /licenses /docker-entrypoint-initdb.d
COPY cockroach.sh cockroach /cockroach/

# Set working directory so that relative paths
# are resolved appropriately when passed as args.
WORKDIR /cockroach/

# Include the directory in the path to make it easier to invoke
# commands via Docker
ENV PATH=/cockroach:$PATH

ENV COCKROACH_CHANNEL=official-docker

EXPOSE 26257 8080
ENTRYPOINT ["/cockroach/cockroach.sh"]
