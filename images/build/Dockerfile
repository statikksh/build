FROM node:lts-buster

WORKDIR /www/statikk/sh

# Updates the system
RUN apt-get upgrade -y && apt-get update -y

# Copies all Python scripts
COPY scripts scripts

# Creates a new user to run the build script
RUN useradd --home-dir /www/statikk/sh statikk && \
    chown -hR statikk /www/statikk/sh && \
    chmod -R 755 /www/statikk/sh

USER statikk

# Runs the builder
CMD [ "python3", "-u", "scripts/build.py" ]