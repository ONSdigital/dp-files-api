Feature: DELETE /files/{filepath}

  Scenario: Successful request returns 204
    Given I am a publisher user
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
      | State         | UPLOADED                                                                  |
    When the file upload "images/with-bundle.jpg" is removed
    Then the HTTP status code should be "204"
    And the file event should be created in the database

  Scenario: Unauthorised request returns 401
    Given I am not identified
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
      | State         | UPLOADED                                                                  |
    When the file upload "images/with-bundle.jpg" is removed
    Then the HTTP status code should be "401"

  Scenario: Forbidden request returns 403 (user does not have required permissions)
    Given I am a JWT user with email "viewer2@ons.gov.uk" and group "role-viewer-denied"
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
      | State         | UPLOADED                                                                  |
    When the file upload "images/with-bundle.jpg" is removed
    Then the HTTP status code should be "403"

  Scenario: Removing a non-existent file returns 404
    Given I am a publisher user
    When the file upload "images/non-existent.jpg" is removed
    Then the HTTP status code should be "404"
  
  Scenario: Removing a "MOVED" file returns 409 (file is already published)
      Given I am a publisher user
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
        | State         | MOVED                                                                     |
      When the file upload "images/with-bundle.jpg" is removed
      Then the HTTP status code should be "409"
      And the file event should be created in the database