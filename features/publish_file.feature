Feature: Publishing file to Kafka from a collection ID

  Scenario: Publishing a collection with a single file associated with it
    Given the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When I publish the collection "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "201"
    And the following document entry should be look like:
      | Path              | images/meme.jpg                                                           |
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | LastModified      | 2021-10-19T09:30:30Z                                                      |
      | PublishedAt       | 2021-10-19T09:30:30Z                                                      |
      | Etag              | 123456789                                                                 |
      | State             | PUBLISHED                                                                 |

  Scenario: Publishing a collection with a multiple file associated with it
    Given the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    Given the file upload "images/other-meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When I publish the collection "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "201"
    And the following document entry should be look like:
      | Path              | images/meme.jpg                                                           |
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | LastModified      | 2021-10-19T09:30:30Z                                                      |
      | PublishedAt       | 2021-10-19T09:30:30Z                                                      |
      | Etag              | 123456789                                                                 |
      | State             | PUBLISHED                                                                 |
    And the following document entry should be look like:
      | Path              | images/other-meme.jpg                                                     |
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | LastModified      | 2021-10-19T09:30:30Z                                                      |
      | PublishedAt       | 2021-10-19T09:30:30Z                                                      |
      | Etag              | 123456789                                                                 |
      | State             | PUBLISHED                                                                 |

  Scenario: Failing to publish a collection with a single file that has not completely uploaded
    Given the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I publish the collection "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "409"

  Scenario: Failing to publish a collection with a multiple files one of which has not completely uploaded
    Given the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    Given the file upload "images/other-meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When I publish the collection "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "409"
    And the following document entry should be look like:
      | Path              | images/other-meme.jpg                                                     |
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | Etag              | 123456789                                                                 |
      | State             | UPLOADED                                                                  |


