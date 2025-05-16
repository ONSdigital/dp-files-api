Feature: Remove File

    As a file upload service
    I want to remove a file
    So that I can delete files that are no longer needed

    Scenario: The one where the removal of the file is not successful
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
        | State         | MOVED                                                                     |
      When the file upload "images/with-bundle.jpg" is removed
      Then the HTTP status code should be "409"

  Scenario: The one where the removal of the file is successful
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
      | State         | UPLOADED                                                                  |
    When the file upload "images/with-bundle.jpg" is removed
    Then the HTTP status code should be "204"