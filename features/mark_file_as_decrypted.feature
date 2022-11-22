Feature: Mark files as decrypted

  As a file publishing service
  I want to register that I have completed the decryption of a file
  So that download services know it is now available for redirection to S3

  Scenario: The one where marking the state as decrypted is successful
    Given I am an authorised user
    And the file upload "index.html" has been published with:
      | Path              | index.html                                                           |
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
      | Etag              | ed1fd569c2a0c3797627cd2c6b03119d                                                                 |
      | State             | PUBLISHED                                                                 |
    When the file "index.html" is marked as decrypted with etag "987654321"
    Then the HTTP status code should be "200"
    And the following document entry should be look like:
      | Path              | index.html                                                           |
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:13:14Z                                                      |
      | PublishedAt       | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-19T09:30:30Z                                                      |
      | DecryptedAt       | 2021-10-19T09:30:30Z                                                      |
      | Etag              | 987654321                                                                 |
      | State             | DECRYPTED                                                                       |

  Scenario: The one where the file is not in PUBLISHED state
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
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
    When the file "images/meme.jpg" is marked as decrypted with etag "987654321"
    Then the HTTP status code should be "409"

  Scenario: The one where the file does not exist
    Given I am an authorised user
    When the file "images/not-found.jpg" is marked as decrypted with etag "987654321"
    Then the HTTP status code should be "404"

  Scenario: The one where the user is not authorised to mark a file as decrypted
    Given I am not an authorised user
    And the file upload "images/meme.jpg" has been published with:
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


