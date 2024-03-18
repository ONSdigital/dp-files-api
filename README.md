# DP Files API

## Introduction
The Files API is part of the [Static Files System](https://github.com/ONSdigital/dp-static-files-compose).
This API is responsible for storing the metadata and state of files.

It is used by the [Upload Service](https://github.com/ONSdigital/dp-upload-service) to store the metadata of the file
being uploaded and keep track of the state of the file during uploaded. During upload the state will be CREATED. Once
the full file have been uploaded the upload service should inform this API the upload is complete and the files state will be
moved to UPLOADED.

Any service interesting in the metadata or the state of a file can just the GET endpoints. A single files metadata can
be retrieved by its path or all files in a collection can be retrieved by ID.

The [Download Service](https://github.com/ONSdigital/dp-download-service) uses this API to see whether a file exists and
what state it is in before attempting to serve the file to consumers wishing to access the file.

The API has two end points to publish files. Files can be individually published by PATCHING the state to be PUBLISHED.
It is also to publish all files in a collection in one call by PATCHING /collection/{collection_id}, this can be used to
reduce the number of API calls required to publish a large collection.
Currently, most calls to a publish file will come from the [Zebedee Publisher](https://github.com/ONSdigital/zebedee/blob/ff5d1a23b2bba50dc1ed67b10fbc213972f9ad21/zebedee-cms/src/main/java/com/github/onsdigital/zebedee/model/publishing/Publisher.java#L153)

When a file is published this API sends a message via Kafka to the [Static File Publisher](https://github.com/ONSdigital/dp-static-file-publisher)
that permanently moves the file and inform this API that the file is now moved via an HTTP call.

### REST API

The api is fully documented in [Swagger Docs](swagger.yaml)

**Note:** When using PATCH calls to modify the file metadata you can either send a `collection_id` to set the collection_id on a file
where it is not already sent or change the `state` of a file.

### Metadata

| Field          | Notes                                                                                                          |
|----------------|----------------------------------------------------------------------------------------------------------------|
| path           | The identifier of a file that is stored. Globally unique, and forms part of the bucket/object name when stored |
| is_publishable | This field currently is ignored and has no affect, the file will be published if a publish update is sent!     |
| collection_id  | Optional during upload, must be set for the file to be published                                               |
| title          | Optional                                                                                                       |
| size_in_bytes  | The size of the file                                                                                           |
| type           | mimetype of the file, e.g. "text/csv", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"     |
| licence        | Freetext name of the licence under which the file is made available                                            |
| licence_url    | URL to the license                                                                                             |
| state          | State of the file - CREATED, UPLOADED, PUBLISHED, MOVED                                                    |
| etag           | Cyrptographic hash of the file content                                                                         |

#### Additional Metadata

Additional timestamp data about the file is stored in the database but not exposed via the API. Those fields are:

| Field               |
|---------------------|
| created_at          |
| last_modified       |
| upload_completed_at |
| published_at        |
| moved_at        |


### File States

| State     | Description                                                                                                                 |
|-----------|-----------------------------------------------------------------------------------------------------------------------------|
| CREATED   | File upload has started and the metadata has been provide to this API                                                       |
| UPLOADED  | File upload has been completed. The etag for the final file has been provided                                               |
| PUBLISHED | The file has been published (it is available to the public, but is not yet permently moved)                             |
| MOVED | The file has been permanently moved and moved to the public bucket for storage. The public files etag has been provided |

```


 Start     ┌──────────────┐ File      ┌──────────────┐ File       ┌───────────────┐ File       ┌───────────────┐
 Upload    │              │ Uploaded  │              │ Published  │               │ Moved      │               │
      ────►│   CREATED    ├──────────►│   UPLOADED   ├───────────►│   PUBLISHED   ├───────────►│     MOVED     │
           │              │           │              │            │               │            │               │
           └──────────────┘           └──────────────┘            └───────────────┘            └───────────────┘
             File is in a               File is ready               File is available           File is available
             unusable                   for review &                for public download         for public download
             state                      approval                    The stored version          directly from S3
             Can resume upoad           Can be pre-viewed           is moved on-demand          where it is stored                 

```

## Getting started

* Run `make debug`

## Dependencies

* No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable         | Default                  | Description                                                                                                        |
|------------------------------|--------------------------|--------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | :26900                   | The host and port to bind to                                                                                       |
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL         | 30s                      | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| IS_PUBLISHING                | false                    | Whether the service is running in the Publishing domain                                                            |
| PERMISSIONS_API_URL          | http://localhost:25400   | The hostname of the permissions API                                                                                |
| IDENTITY_API_URL             | http://localhost:25600   | The hostname of the identity API                                                                                   |
| ZEBEDEE_URL                  | http://localhost:8082    | The hostname of the zebedee API                                                                                    |
| KAFKA_ADDR                   | `kafka:9092`             | A (comma delimited) list of kafka brokers (TLS-ready)                                                              |
| KAFKA_VERSION                | `2.6.1`                  | The version of (TLS-ready) Kafka being used                                                                        |
| KAFKA_MAX_BYTES              | `200000`                 | The max message size for kafka producer                                                                            |
| KAFKA_SEC_PROTO              | _unset_                  | if set to `TLS`, kafka connections will use TLS ([ref-1])                                                          |
| KAFKA_SEC_CLIENT_KEY         | _unset_                  | PEM for the client key ([ref-1])                                                                                   |
| KAFKA_SEC_CLIENT_CERT        | _unset_                  | PEM for the client certificate ([ref-1])                                                                           |
| KAFKA_SEC_CA_CERTS           | _unset_                  | CA cert chain for the server cert ([ref-1])                                                                        |
| KAFKA_SEC_SKIP_VERIFY        | false                    | ignores server certificate issues if `true` ([ref-1])                                                              |
| STATIC_FILE_PUBLISHED_TOPIC  | static-file-published-v2 |                                                                                                                    |
| MONGODB_BIND_ADDR            | `localhost:27017`        | Address of MongoDB                                                                                                 |
| MONGODB_DATABASE             | `files`                  | The mongodb database to store imports                                                                              |
| MONGODB_COLLECTIONS          | `metadata`               | The (comma delimited) list of mongodb collections to store imports                                                 |
| MONGODB_USERNAME             | _unset_                  | The mongodb username                                                                                               |
| MONGODB_PASSWORD             | _unset_                  | The mongodb username                                                                                               |
| MONGODB_ENABLE_READ_CONCERN  | false                    | Switch to use (or not) majority read concern                                                                       |
| MONGODB_ENABLE_WRITE_CONCERN | true                     | Switch to use (or not) majority write concern                                                                      |
| MONGODB_CONNECT_TIMEOUT      | 5s                       | The default timeout when connecting to mongodb                                                                     |
| MONGODB_QUERY_TIMEOUT        | 15s                      | The default timeout for querying mongodb                                                                           |
| MONGODB_IS_SSL               | false                    | Switch to use (or not) TLS when connecting to mongodb                                                              |
| MONGODB_VERIFY_CERT          |                          |                                                                                                                    |
| MONGODB_CERT_CHAIN           |                          |                                                                                                                    |
| MONGODB_REAL_HOSTNAME        |                          |                                                                                                                    |


## API Client

There is an [API Client](https://github.com/ONSdigital/dp-api-clients-go/tree/main/files) for the File API this is part
of [dp-api-clients-go](https://github.com/ONSdigital/dp-api-clients-go) package.

The Files Client provides functions that enables:

  - Setting the Collection ID of existing file
  - Publish all Files in a Collection
  - Get the details of a single file

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

