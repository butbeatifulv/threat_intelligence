FROM neo4j:5
USER root
RUN apt-get update && apt-get install -y --no-install-recommends curl unzip ca-certificates python3 \
  && rm -rf /var/lib/apt/lists/*
COPY deploy/knowledge/docker/graph-bootstrap.sh /graph-bootstrap.sh
RUN chmod +x /graph-bootstrap.sh
USER neo4j
ENTRYPOINT ["/graph-bootstrap.sh"]
