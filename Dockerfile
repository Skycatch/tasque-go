FROM golang:1.12.7-stretch as builder

ARG arch
ARG version
ARG git_summary
ARG build_date
ARG git_sha
ARG git_branch

WORKDIR /app
COPY . .
RUN go get && \
  CGO_ENABLED=0 GOOS=linux GOARCH=${arch} go build \
	-a -installsuffix cgo \
	-ldflags "-X 'main.Version=${version}' -X 'main.GitSummary=${git_summary}' -X 'main.BuildDate=${build_date}' -X main.GitCommit=${git_sha} -X main.GitBranch=${git_branch}" \
	-o tasque .
  
FROM centurylink/ca-certs
COPY --from=builder /app/tasque /tasque
ENTRYPOINT ["/tasque"]