
## Environment based configuration

* `S3_ENDPOINT` default: `"s3.amazonaws.com"`
    * host:port address.
* `S3_ACCESS_KEY_ID`
    * Access Key ID for the service.
* `S3_SECRET_ACCESS_KEY`
    * Secret Access Key for the service.
* `S3_USE_SSL`
    * Determine whether to use SSL or not.
* `S3_REGION` default: `"us-east-1"`
    * S3 region for the bucket.
* `S3_HEALTH_CHECK_BUCKET_NAME` default: `health-check`
    * Name of the bucket that is used for health check operations.
* `S3_HEALTH_CHECK_OBJECT_NAME` default: `"latest.log`
    * Name of the object that is used for the health check operation.
* `S3_HEALTH_CHECK_RESULT_TTL` default: `10s`
    * Amount of time to cache the last health check result.