FROM scratch
COPY tray /usr/bin/tray
ENV HOME=/home/user
# So -v ~/.tray:/home/user/.tray uses the same key as host.
VOLUME ["/home/user/.tray"]
ENTRYPOINT ["/usr/bin/tray"]
