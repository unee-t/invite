FROM scratch
COPY invite /
WORKDIR /
ENV PORT 9000
ENTRYPOINT ["/invite"]
