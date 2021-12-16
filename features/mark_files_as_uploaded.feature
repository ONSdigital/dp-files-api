Feature: Mark files as uploaded

    As a file upload service
    I want to register that i have complete upload of a file
    So that download services know it is now available

    Scenario: The one where marking upload complete is successful
        Given the file upload "/images/meme.jpg" has been registered with:
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
        When the file upload "/images/meme.jpg" is marked as complete with the etag "123456789"
        Then the HTTP status code should be "201"
        And the following document entry should be look like:
            | Path              | /images/meme.jpg                                                          |
            | IsPublishable     | true                                                                      |
            | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
            | Title             | The latest Meme                                                           |
            | SizeInBytes       | 14794                                                                     |
            | Type              | image/jpeg                                                                |
            | Licence           | OGL v3                                                                    |
            | LicenceUrl        | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
            | CreatedAt         | 2021-10-21T15:13:14Z                                                      |
            | LastModified      | 2021-10-19T09:30:30Z                                                      |
            | UploadCompletedAt | 2021-10-19T09:30:30Z                                                      |
            | Etag              | 123456789                                                                 |
            | State             | UPLOADED                                                                  |

#  Scenario: The one where the DB is down
#
    Scenario: Trying to mark an upload complete on a file that was not registered
        Given the file upload "/images/meme.jpg" has not been registered
        When the file upload "/images/meme.jpg" is marked as complete with the etag "123456789"
        Then the HTTP status code should be "404"

    Scenario: Trying to mark an upload complete on a file that is in the wrong state
        Given the file upload "/images/meme.jpg" has been registered with:
            | IsPublishable | true                                                                      |
            | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
            | Title         | The latest Meme                                                           |
            | SizeInBytes   | 14794                                                                     |
            | Type          | image/jpeg                                                                |
            | Licence       | OGL v3                                                                    |
            | LicenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
            | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
            | LastModified  | 2021-10-21T15:13:14Z                                                      |
            | State         | UPLOADED                                                                  |
        When the file upload "/images/meme.jpg" is marked as complete with the etag "123456789"
        Then the HTTP status code should be "409"