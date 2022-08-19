FROM golang:1.16
WORKDIR /exporter
COPY go.mod go.sum ./
# Download the dependencies
RUN go mod download
RUN useradd exporter-user
#Copy source file
COPY main.go ./
RUN go build -o s3_bucket_exporter main.go
USER exporter-user
EXPOSE 20000

CMD [ "/exporter/s3_bucket_exporter" ]