Feature: Creating file events

  Scenario: Successfully create a file event when authorised
    Given I am an authorised user
    When I create a file event with payload:
        """
        {
          "requested_by": {
            "id": "user123",
            "email": "user@example.com"
          },
          "action": "READ",
          "resource": "/downloads/test-file.csv",
          "file": {
            "path": "test-file.csv",
            "is_publishable": true,
            "title": "Test File",
            "size_in_bytes": 1024,
            "type": "text/csv",
            "licence": "OGL v3",
            "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
          }
        }
        """
    Then the HTTP status code should be "201"
    And the file event should be created in the database

  Scenario: Cannot create a file event when not authorised
    Given I am not an authorised user
    When I create a file event with payload:
        """
        {
          "requested_by": {
            "id": "user123"
          },
          "action": "READ",
          "resource": "/downloads/test-file.csv",
          "file": {
            "path": "test-file.csv",
            "type": "text/csv"
          }
        }
        """
    Then the HTTP status code should be "403"

  Scenario: Error when creating a file event with invalid JSON
    Given I am an authorised user
    When I create a file event with payload:
        """
        { invalid json }
        """
    Then the HTTP status code should be "400"

  Scenario: Create file event for bundle file download
    Given I am an authorised user
    When I create a file event with payload:
        """
        {
          "requested_by": {
            "id": "user789",
            "email": "downloader@example.com"
          },
          "action": "READ",
          "resource": "/downloads/bundle-file.csv",
          "file": {
            "path": "bundle-file.csv",
            "is_publishable": false,
            "bundle_id": "bundle-123",
            "title": "Bundle Data File",
            "size_in_bytes": 2048,
            "type": "text/csv"
          }
        }
        """
    Then the HTTP status code should be "201"
    And the file event should be created in the database