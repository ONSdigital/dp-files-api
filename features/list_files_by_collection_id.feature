Feature: List Files by Collection ID

  Scenario: The one where there are no files in the collection
    Given I am an authorised user
    When I get files in the collection "1234-asdfg-54321-qwerty"
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

  Scenario: The one where there are some file in the collection
    Given I am an authorised user
    Given the file upload "images/meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
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
    And the file upload "images/other-meme.jpg" has been completed with:
      | IsPublishable     | true                                                                      |
      | CollectionID      | 1234-asdfg-54321-qwerty                                                   |
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
    When I get files in the collection "1234-asdfg-54321-qwerty"
    Then I should receive the following JSON response with status "200":
    """
{
  "count": 2,
  "limit": 2,
  "offset": 0,
  "total_count": 2,
  "items": [
    {
      "path": "images/meme.jpg",
      "is_publishable": true,
      "collection_id": "1234-asdfg-54321-qwerty",
      "title": "The latest Meme",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "UPLOADED",
      "etag": "123456789"
    },
    {
      "path": "images/other-meme.jpg",
      "is_publishable": true,
      "collection_id": "1234-asdfg-54321-qwerty",
      "title": "The latest Meme",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "UPLOADED",
      "etag": "123456789"
    }
  ]
}
    """

  Scenario: The one where the user is not authorised to view a list of files
    Given I am not an authorised user
    When I get files in the collection "1234-asdfg-54321-qwerty"
    Then the HTTP status code should be "403"
