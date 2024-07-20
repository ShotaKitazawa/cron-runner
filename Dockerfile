FROM gcr.io/distroless/static:nonroot
COPY cron-runner /
USER 65532:65532
ENTRYPOINT ["/cron-runner"]
