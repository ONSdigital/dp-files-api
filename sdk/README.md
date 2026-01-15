# dp-files-api SDK

## Overview

This SDK provides a client for interacting with the dp-files-api. It is intended to be consumed by services that require endpoints from the dp-files-api. It also provides healthcheck functionality, mocks and structs for easy integration, testing and error handling.

## Available client methods

| Name | Description |
| ------ | ------------- |
| [`Checker`](#checker) | Calls the `health.Client`'s `Checker` method |
| [`Health`](#health) | Returns the underlying Healthcheck Client for this API client |
| [`URL`](#url) | Returns the URL used by this client |
| [`DeleteFile`](#deletefile) | Deletes a file at the specified filePath |
| [`CreateFileEvent`](#createfileevent) | Creates a new file event in the audit log and returns the created event |
| [`GetFile`](#getfile) | Retrieves the metadata for a file at the specified path |
| [`MarkFilePublished`](#markfilepublished) | Sets the state of a file to `PUBLISHED` |

## Instantiation

Example using `New`:

```go
package main

import "github.com/ONSdigital/dp-files-api/sdk"

func main() {
    client := sdk.New("http://localhost:26900")
}
```

Example using `NewWithHealthClient`:

```go
package main

import (
    "github.com/ONSdigital/dp-api-clients-go/v2/health"
    "github.com/ONSdigital/dp-files-api/sdk"
)

func main() {
    existingHealthClient := health.NewClient("existing-service-name", "http://localhost:8080")

    client := sdk.NewWithHealthClient(existingHealthClient)
}
```

## Example usage of client

This example demonstrates how the `GetFile()` function could be used:

```go
package main

import (
    "context"

    "github.com/ONSdigital/dp-files-api/sdk"
)

func main() {
    client := sdk.New("http://localhost:26900")

    headers := sdk.Headers{
        Authorization: "auth-token",
    }

    fileMetadata, err := client.GetFile(context.Background(), "/path/to/file.csv", headers)
    if err != nil {
        // Distinguish between API errors and other errors
        apiErr, ok := err.(*sdk.APIError)
        if ok {
            // Type is *sdk.APIError so we can access all the following fields:
            // apiErr.StatusCode
            // apiErr.Errors // This is an array that can be looped through
            // apiErr.Errors.Error[0].Code
            // apiErr.Errors.Error[0].Description
            // apiErr.Error()
        } else {
            // Handle non-API errors
        }
    }
}
```

## Available Functionality

### Checker

```go
import "github.com/ONSdigital/dp-healthcheck/healthcheck"

check := &healthcheck.CheckState{}
err := client.Checker(ctx, check)
```

### Health

```go
healthClient := client.Health()
```

### URL

```go
url := client.URL()
```

### DeleteFile

```go
err := client.DeleteFile(ctx, "/path/to/delete.csv", sdk.Headers{})
```

### CreateFileEvent

```go
import "github.com/ONSdigital/dp-files-api/files"

fileEvent := files.FileEvent{
    RequestedBy: &files.RequestedBy{
        ID:    "user",
        Email: "user@email.com",
    },
    Action:   files.ActionRead,
    Resource: "/path/to/file.csv",
    File: &files.FileMetaData{
        Path: "/path/to/file.csv",
        // Add additional fields
    },
}

createdFileEvent, err := client.CreateFileEvent(ctx, fileEvent, sdk.Headers{})
```

### GetFile

```go
fileMetadata, err := client.GetFile(ctx, "/path/to/file.csv", sdk.Headers{})
```

### MarkFilePublished

```go
err := client.MarkFilePublished(ctx, "/path/to/file.csv", sdk.Headers{})
```

## Additional Information

### Errors

The [`APIError`](errors.go) struct allows the user to distinguish if an error is a generic error or an API error, therefore allowing access to more detailed fields. This is shown in the [Example usage of client](#example-usage-of-client) section.

### Headers

The [`Headers`](headers.go) struct allows the user to provide an Authorization header if required.
This must be set without the `"Bearer "` prefix as the SDK will automatically add this.

### Mocks

To simplify testing, all functions provided by the client have been defined in the [`Clienter` interface](interface.go). This allows the user to use [auto-generated mocks](mocks/) within unit tests.

Example of how to define a mock clienter:

```go
import (
    "context"
    "testing"

    "github.com/ONSdigital/dp-files-api/files"
    "github.com/ONSdigital/dp-files-api/sdk/mocks"
)

func Test(t *testing.T) {
    mockClient := mocks.ClienterMock{
        GetFileFunc: func(ctx context.Context, filePath string) (*files.StoredRegisteredMetaData, error) {
            // Setup mock behaviour here
            return &files.StoredRegisteredMetaData{}, nil
        },
        // Other methods can be mocked if needed
    }
}
```
