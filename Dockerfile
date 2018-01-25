FROM scratch
COPY almanac-linux-static .
ENTRYPOINT ["./almanac-linux-static"]

