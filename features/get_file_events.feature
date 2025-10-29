Feature: Getting file events

  Background:
    Given I am an authorised user

  Scenario: Successfully get all file events with full JSON response
    Given the following file events exist in the database:
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
          "created_at": "2025-10-28T10:00:00Z",
          "requested_by": {
            "id": "user456",
            "email": "user456@example.com"
          },
          "action": "READ",
          "resource": "/downloads/file2.csv",
          "file": {
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
            "path": "file3.csv",
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
          "created_at": "2025-10-28T07:00:00Z",
          "requested_by": {
            "id": "user456",
            "email": "user456@example.com"
          },
          "action": "DELETE",
          "resource": "/files/file2.csv",
          "file": {
            "path": "file2.csv",
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
    Given the following file events exist in the database:
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
    Given the following file events exist in the database:
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
            "path": "file1.csv",
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
    Given the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
      | user456       | CREATE | /files/file2.csv          | file2.csv |
    When I GET "/file-events?after=2020-01-01T00:00:00Z"
    Then the HTTP status code should be "200"
    And the response should contain at least "1" file event

  Scenario: Cannot get file events when not authorised
    Given I am not an authorised user
    When I GET "/file-events"
    Then the HTTP status code should be "403"

  Scenario: Return 404 when filtering by non-existent path
    Given the following file events exist in the database:
      | RequestedByID | Action | Resource                  | FilePath  |
      | user123       | READ   | /downloads/file1.csv      | file1.csv |
    When I GET "/file-events?path=nonexistent-file.csv"
    Then the HTTP status code should be "404"

  Scenario: Return 400 for invalid limit parameter
    When I GET "/file-events?limit=abc"
    Then the HTTP status code should be "400"

  Scenario: Return 400 for invalid offset parameter
    When I GET "/file-events?offset=-5"
    Then the HTTP status code should be "400"

  Scenario: Return 400 for invalid date format
    When I GET "/file-events?after=not-a-date"
    Then the HTTP status code should be "400"

  Scenario: Return 400 for limit exceeding maximum
    When I GET "/file-events?limit=2000"
    Then the HTTP status code should be "400"

  Scenario: Get empty results when no events match filters
    Given the following file events exist in the database:
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
