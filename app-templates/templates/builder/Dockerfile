FROM docker.io/library/golang:1.27.2-bookworm
RUN apt-get update && apt-get install -y --no-install-recommends openssh-client make gawk \
    && rm -rf /var/lib/apt/lists/*
