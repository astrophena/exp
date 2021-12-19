FROM gcr.io/distroless/base-debian11
COPY sqlplay /usr/local/bin/sqlplay
COPY chinook.db /usr/local/share/chinook.db
CMD ["/usr/local/bin/sqlplay", "-addr", ":3000", "-db", "/usr/local/share/chinook.db"]
