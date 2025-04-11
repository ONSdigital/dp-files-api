Feature: Updating a file's bundle ID

  Scenario: Setting a bundle ID for a file
    Given I am an authorised user
    And the file upload "images/bundleless.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | Image without bundle                                                      |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I set the bundle ID to "new-bundle-456" for file "images/bundleless.jpg"
    Then the HTTP status code should be "200"
    When the file metadata is requested for the file "images/bundleless.jpg"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "images/bundleless.jpg",
      "is_publishable": true,
      "bundle_id": "new-bundle-456",
      "title": "Image without bundle",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    """

  Scenario: Error when setting a bundle ID for a file that already has one
    Given I am an authorised user
    And the file upload "images/with-bundle.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | BundleID      | existing-bundle-789                                                       |
      | Title         | Image with existing bundle                                                |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I set the bundle ID to "another-bundle-999" for file "images/with-bundle.jpg"
    Then the HTTP status code should be "400"
    
  Scenario: Error when setting a bundle ID for a file that has not been registered
    Given I am an authorised user
    And the file "images/not-registered.jpg" has not been registered
    When I set the bundle ID to "any-bundle-id" for file "images/not-registered.jpg"
    Then the HTTP status code should be "404"
    
  Scenario: Error when the user is not authorised to update a file
    Given I am not an authorised user
    And the file upload "images/unauthorised.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | Unauthorised image                                                        |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I set the bundle ID to "forbidden-bundle" for file "images/unauthorised.jpg"
    Then the HTTP status code should be "403"

Scenario: Removing a bundle ID from a file
  Given I am an authorised user
  And the file upload "images/with-bundle.jpg" has been registered with:
    | IsPublishable | true                                                                      |
    | BundleID      | existing-bundle-789                                                       |
    | Title         | Image with existing bundle                                                |
    | SizeInBytes   | 14794                                                                     |
    | Type          | image/jpeg                                                                |
    | Licence       | OGL v3                                                                    |
    | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
    | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
    | LastModified  | 2021-10-21T15:13:14Z                                                      |
    | State         | CREATED                                                                   |
  When I set the bundle ID to "" for file "images/with-bundle.jpg"
  Then the HTTP status code should be "200"
  When the file metadata is requested for the file "images/with-bundle.jpg"
  Then I should receive the following JSON response with status "200":
  """
  {
    "path": "images/with-bundle.jpg",
    "is_publishable": true,
    "title": "Image with existing bundle",
    "size_in_bytes": 14794,
    "type": "image/jpeg",
    "licence": "OGL v3",
    "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
    "state": "CREATED",
    "etag": ""
  }
  """