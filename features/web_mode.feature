Feature: Web mode restrictions

  Background:
    Given I am in web mode

  Scenario: The one where I try to register a new file
    When the file upload is registered with payload:
        """
        {}
        """
    Then the HTTP status code should be "403"

  Scenario: The One where I try to mark a file as uploaded
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
    When the file upload "images/meme.jpg" is marked as complete with the etag "123456789"
    Then the HTTP status code should be "403"

  Scenario: The one where I try to mark a file as published
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
    Then the HTTP status code should be "403"

  Scenario: The one where I try to mark the file as decrypted
    Given the file upload "images/meme.jpg" has been published with:
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
      | LastModified      | 2021-10-21T15:13:14Z                                                      |
      | PublishedAt       | 2021-10-21T15:13:14Z                                                      |
      | Etag              | 123456789                                                                 |
      | State             | PUBLISHED                                                                 |
    When the file "images/meme.jpg" is marked as decrypted with etag "987654321"
    Then the HTTP status code should be "403"

  Scenario: The one where I try to get a file
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
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "200"
