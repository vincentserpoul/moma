FROM centurylink/ca-certs
EXPOSE 8080 80 9000
COPY moma /
ENTRYPOINT ["/moma"]
