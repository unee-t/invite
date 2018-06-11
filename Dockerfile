FROM scratch
COPY invite /
WORKDIR /
COPY sql/ /sql/
ENV PORT 9000
ENTRYPOINT ["/invite"]
