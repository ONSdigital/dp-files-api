Feature: Mark single file as published

  Scenario: The one where marking the state as published is successful
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been completed with:
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
    And Kafka Consumer Group is running
    When the file "images/meme.jpg" is marked as published
    Then the HTTP status code should be "200"
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
    And the following PUBLISHED message is sent to Kakfa:
      | path        | images/meme.jpg |
      | etag        | 123456789       |
      | type        | image/jpeg      |
      | sizeInBytes | 14794           |

  Scenario: The one where marking the state as published is invalid state move
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | Title             | The latest Meme                                                           |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | PUBLISHED                                                                 |
      | Etag              | 123456789                                                                 |
    When the file "images/meme.jpg" is marked as published
    Then the HTTP status code should be "409"

  Scenario: The one where the collection ID is not set
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
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
    When the file "images/meme.jpg" is marked as published
    Then the HTTP status code should be "409"

  Scenario: The one where the file does not exists
    Given I am an authorised user
    When the file "images/meme.jpg" is marked as published
    Then the HTTP status code should be "404"

  Scenario: The one where the file is not publishable
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | false                                                                     |
      | Title             | The latest Meme                                                           |
      | SizeInBytes       | 14794                                                                     |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Type              | image/jpeg                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When the file "images/meme.jpg" is marked as published
    Then the HTTP status code should be "409"

  Scenario: The one where the user is not authorised to mark a file as published
    Given I am not an authorised user
    And the file upload "images/meme.jpg" has been completed with:
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
    When the file "images/meme.jpg" is marked as published
    Then the HTTP status code should be "403"

  Scenario: The one where the file name is a bit weird
    Given I am an authorised user
    And the file upload "interactives/87a3dde3-wéî®∂-4290-9a3b-afbea82e0fa7/version-11/lib&/chosen-sprite@2x.png" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
      | Title             | Stranger Files                                                          |
      | SizeInBytes       | 14794                                                                     |
      | Type              | image/png                                                                |
      | Licence           | OGL v3                                                                    |
      | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
      | LastModified      | 2021-10-21T15:14:14Z                                                      |
      | UploadCompletedAt | 2021-10-21T15:14:14Z                                                      |
      | State             | UPLOADED                                                                  |
      | Etag              | 123456789                                                                 |
    When the file "interactives/87a3dde3-wéî®∂-4290-9a3b-afbea82e0fa7/version-11/lib&/chosen-sprite@2x.png" is marked as published
    Then the HTTP status code should be "200"
