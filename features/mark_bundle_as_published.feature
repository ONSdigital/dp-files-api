Feature: Publishing file to Kafka from a bundle ID

  Scenario: Failing to publish a bundle with a single file that has not completely uploaded
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | BundleID      | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I publish the bundle "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "409"

  Scenario: Failing to publish a bundle with a multiple files one of which has not completely uploaded
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | BundleID      | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    Given the file upload "images/other-meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | BundleID          | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceURL        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When I publish the bundle "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "409"
    And the following document entry should look like:
      | Path              | images/other-meme.jpg                                                     |
      | IsPublishable     | true                                                                      |
      | BundleID          | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceURL        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | Etag              | 123456789                                                                 |
      | State             | UPLOADED                                                                  |

  Scenario: Publishing file for a bundle that does not exists
    Given I am an authorised user
    When I publish the bundle "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "201"

  Scenario: The one where the user is not authorised to publish a bundle
    Given I am not an authorised user
    And the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | BundleID          | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceURL        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When I publish the bundle "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "403"
