# S3 Bucket Exporter

A very simple check to count the amount of files added to a bucket and path.

## Build the executable

```
go build .
```

# Run it as a docker container

Build the container

```
docker build --tag s3_bucket_exporter .
```

Run the container locally

```
docker run s3_bucket_exporter /exporter/s3_bucket_exporter
```