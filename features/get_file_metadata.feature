Feature: Fetching metadata for a file

  Scenario: The file metadata is retrieved when file upload has been registered
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "images/meme.jpg",
      "is_publishable": true,
      "collection_id": "1234-asdfg-54321-qwerty",
      "title": "The latest Meme",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    """

  Scenario: The one where the user is not authorised to (pre)view
    Given I am not an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "403"


  Scenario: The one where the collection ID is not set
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "images/meme.jpg",
      "is_publishable": true,
      "title": "The latest Meme",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    """

  Scenario: The file metadata is not found when file has not been registered
    Given I am an authorised user
    And the file "images/not-found.jpg" has not been registered
    When the file metadata is requested for the file "images/not-found.jpg"
    Then the HTTP status code should be "404"
    
  Scenario: The file metadata is retrieved with bundle ID
    Given I am an authorised user
    And the file upload "images/bundle-image.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | BundleID      | bundle-123                                                                |
      | Title         | Bundle Image                                                              |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/bundle-image.jpg"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "images/bundle-image.jpg",
      "is_publishable": true,
      "bundle_id": "bundle-123",
      "title": "Bundle Image",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    """
  
  Scenario: Retrieve file metadata with content_item
    Given I am an authorised user
    And the file upload "datasets/cpih/2024/data.csv" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | CPIH Dataset 2024                                                         |
      | SizeInBytes   | 54321                                                                     |
      | Type          | text/csv                                                                  |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | DatasetID     | cpih-dataset-001                                                          |
      | Edition       | 2024                                                                      |
      | Version       | 1                                                                         |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "datasets/cpih/2024/data.csv"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "datasets/cpih/2024/data.csv",
      "is_publishable": true,
      "title": "CPIH Dataset 2024",
      "size_in_bytes": 54321,
      "type": "text/csv",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "content_item": {
        "dataset_id": "cpih-dataset-001",
        "edition": "2024",
        "version": "1"
      },
      "state": "CREATED",
      "etag": ""
    }
    """

  Scenario: File metadata request without a JWT returns 401
    Given I am not authenticated
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "401"

  Scenario: File metadata request with an invalid JWT returns 401
    Given I set the "Authorization" header to "Bearer invalid-token"
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "401"

  Scenario: File metadata request with a service token returns 401
    Given I use a service auth token "service-token"
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "401"

  Scenario: File metadata request with insufficient permissions returns 403
    Given I am not an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "403"
