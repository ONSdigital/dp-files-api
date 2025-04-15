Feature: List Files by Bundle ID

  Scenario: The one where there are no files in the bundle
    Given I am an authorised user
    When I get files in the bundle "non-existent-bundle"
    Then I should receive the following JSON response with status "200":
    """
    {
    "count": 0,
    "limit": 0,
    "offset": 0,
    "total_count": 0,
    "items": []
    }
    """

  Scenario: The one where there are some files in the bundle
    Given I am an authorised user
    Given the file upload "images/bundle1-image1.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | BundleID      | test-bundle-1                                                             |
      | Title         | Bundle 1 Image 1                                                          |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    And the file upload "images/bundle1-image2.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | BundleID      | test-bundle-1                                                             |
      | Title         | Bundle 1 Image 2                                                          |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I get files in the bundle "test-bundle-1"
    Then I should receive the following JSON response with status "200":
    """
    {
    "count": 2,
    "limit": 2,
    "offset": 0,
    "total_count": 2,
    "items": [
    {
      "path": "images/bundle1-image1.jpg",
      "is_publishable": true,
      "bundle_id": "test-bundle-1",
      "title": "Bundle 1 Image 1",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    },
    {
      "path": "images/bundle1-image2.jpg",
      "is_publishable": true,
      "bundle_id": "test-bundle-1",
      "title": "Bundle 1 Image 2",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    ]
    }
    """

  Scenario: The one where the user is not authorised to view a list of files
    Given I am not an authorised user
    When I get files in the bundle "test-bundle-1"
    Then the HTTP status code should be "403"
    
  Scenario: The one where both collection ID and bundle ID are provided
    Given I am an authorised user
    When I get files with both collection_id "coll-1234" and bundle_id "bundle-1234"
    Then the HTTP status code should be "400"