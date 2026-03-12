Feature: Getting file events

  Scenario: Successfully get all file events with full JSON response
    Given I am an admin user
    And the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | READ   | /downloads/file2.csv      | file2.csv |
      | user789       | CREATE | /files/file3.csv          | file3.csv |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | DELETE | /files/file2.csv          | file2.csv |
    When I GET "/file-events"
    Then I should receive the following JSON response with status "200":
    """
    {
      "count": 5,
      "limit": 20,
      "offset": 0,
      "total_count": 5,
      "items": [
        {
          "created_at": "2025-10-28T11:00:00Z",
          "requested_by": {
            "id": "user123",
            "email": "user123@example.com"
          },
          "action": "READ",
          "resource": "/downloads/file1.csv",
          "file": {
            "etag":"",
            "path": "file1.csv",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "state":"",
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        },
        {
          "created_at": "2025-10-28T10:00:00Z",
          "requested_by": {
            "id": "user456",
            "email": "user456@example.com"
          },
          "action": "READ",
          "resource": "/downloads/file2.csv",
          "file": {
            "etag":"",
            "path": "file2.csv",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "state":"",
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        },
        {
          "created_at": "2025-10-28T09:00:00Z",
          "requested_by": {
            "id": "user789",
            "email": "user789@example.com"
          },
          "action": "CREATE",
          "resource": "/files/file3.csv",
          "file": {
            "etag":"",
            "path": "file3.csv",
            "state":"",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        },
        {
          "created_at": "2025-10-28T08:00:00Z",
          "requested_by": {
            "id": "user123",
            "email": "user123@example.com"
          },
          "action": "READ",
          "resource": "/downloads/file1.csv",
          "file": {
            "etag":"",
            "path": "file1.csv",
            "state":"",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        },
        {
          "created_at": "2025-10-28T07:00:00Z",
          "requested_by": {
            "id": "user456",
            "email": "user456@example.com"
          },
          "action": "DELETE",
          "resource": "/files/file2.csv",
          "file": {
            "etag":"",
            "path": "file2.csv",
            "state":"",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        }
      ]
    }
    """

  Scenario: Get file events with pagination and full JSON response
    Given I am an admin user
    And the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | READ   | /downloads/file2.csv      | file2.csv |
      | user789       | CREATE | /files/file3.csv          | file3.csv |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | DELETE | /files/file2.csv          | file2.csv |
    When I GET "/file-events?limit=2&offset=1"
    Then I should receive the following JSON response with status "200":
    """
    {
      "count": 2,
      "limit": 2,
      "offset": 1,
      "total_count": 5,
      "items": [
        {
          "created_at": "2025-10-28T10:00:00Z",
          "requested_by": {
            "id": "user456",
            "email": "user456@example.com"
          },
          "action": "READ",
          "resource": "/downloads/file2.csv",
          "file": {
            "etag":"",
            "state":"",
            "path": "file2.csv",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        },
        {
          "created_at": "2025-10-28T09:00:00Z",
          "requested_by": {
            "id": "user789",
            "email": "user789@example.com"
          },
          "action": "CREATE",
          "resource": "/files/file3.csv",
          "file": {
            "etag":"",
            "state":"",
            "path": "file3.csv",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        }
      ]
    }
    """

  Scenario: Get file events filtered by path with full JSON response
    Given I am an admin user
    And the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | READ   | /downloads/file2.csv      | file2.csv |
      | user789       | CREATE | /files/file1.csv          | file1.csv |
    When I GET "/file-events?path=file1.csv"
    Then I should receive the following JSON response with status "200":
    """
    {
      "count": 2,
      "limit": 20,
      "offset": 0,
      "total_count": 2,
      "items": [
        {
          "created_at": "2025-10-28T11:00:00Z",
          "requested_by": {
            "id": "user123",
            "email": "user123@example.com"
          },
          "action": "READ",
          "resource": "/downloads/file1.csv",
          "file": {
            "etag":"",
            "state":"",
            "path": "file1.csv",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        },
        {
          "created_at": "2025-10-28T09:00:00Z",
          "requested_by": {
            "id": "user789",
            "email": "user789@example.com"
          },
          "action": "CREATE",
          "resource": "/files/file1.csv",
          "file": {
            "etag":"",
            "path": "file1.csv",
            "state":"",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        }
      ]
    }
    """

  Scenario: Get file events with date filter
    Given I am an admin user
    And the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | CREATE | /files/file2.csv          | file2.csv |
    When I GET "/file-events?after=2020-01-01T00:00:00Z"
    Then the HTTP status code should be "200"
    And the response should contain at least "1" file event

  Scenario: Cannot get file events when not authorised
    Given I am not identified
    When I GET "/file-events"
    Then the HTTP status code should be "401"

  Scenario: Return 404 when filtering by non-existent path
    Given I am an admin user
    And the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
    When I GET "/file-events?path=nonexistent-file.csv"
    Then the HTTP status code should be "404"

  Scenario: Return 400 for invalid limit parameter
    Given I am an admin user
    When I GET "/file-events?limit=abc"
    Then I should receive the following JSON response with status "400":
    """
    {
      "errors": [
        {
          "errorCode": "InvalidRequest",
          "description": "unable to process request due to a malformed or invalid request body or query parameter"
        }
      ]
    }
    """

  Scenario: Return 400 for invalid offset parameter
    Given I am an admin user
    When I GET "/file-events?offset=-5"
    Then I should receive the following JSON response with status "400":
    """
    {
      "errors": [
        {
          "errorCode": "InvalidRequest",
          "description": "unable to process request due to a malformed or invalid request body or query parameter"
        }
      ]
    }
    """

  Scenario: Return 400 for invalid date format
    Given I am an admin user
    When I GET "/file-events?after=not-a-date"
    Then I should receive the following JSON response with status "400":
    """
    {
      "errors": [
        {
          "errorCode": "InvalidRequest",
          "description": "unable to process request due to a malformed or invalid request body or query parameter"
        }
      ]
    }
    """

  Scenario: Return 400 for limit exceeding maximum
    Given I am an admin user
    When I GET "/file-events?limit=2000"
    Then I should receive the following JSON response with status "400":
    """
    {
      "errors": [
        {
          "errorCode": "InvalidRequest",
          "description": "unable to process request due to a malformed or invalid request body or query parameter"
        }
      ]
    }
    """

  Scenario: Get empty results when no events match filters
    Given I am an admin user
    And the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
    When I GET "/file-events?path=file1.csv&after=2050-01-01T00:00:00Z"
    Then I should receive the following JSON response with status "200":
    """
    {
      "count": 0,
      "limit": 20,
      "offset": 0,
      "total_count": 0,
      "items": []
    }
    """

  Scenario: A READ audit event is created when file events are successfully retrieved
  Given I am an admin user
  And the following file events exist in the database:
    | RequestedByID | Action | Resource             | FilePath  |
    | user123       | READ   | /downloads/file1.csv | file1.csv |
  When I GET "/file-events"
  Then the HTTP status code should be "200"
  And a READ audit event should be created for the file-events endpoint
