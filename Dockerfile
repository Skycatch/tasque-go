FROM centurylink/ca-certs
COPY tasque /tasque
CMD ["/tasque"]
