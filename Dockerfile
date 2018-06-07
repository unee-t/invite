FROM scratch
COPY invite /
ENV PORT 9000
ENTRYPOINT ["/invite"]
