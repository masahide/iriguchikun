FROM gcr.io/distroless/base
COPY /iriguchikun /iriguchikun
ENTRYPOINT ["/iriguchikun"]
