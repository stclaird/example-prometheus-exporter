# S3 Bucket Exporter

A very simple example prometheus exporter that will  count the amount of files added to a particular S3 bucket and path. You might use this to check an uploads directory, to ensure that files are being processed correctly.   

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