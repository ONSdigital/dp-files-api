Feature: Fetching metadata for a file

  Scenario: The file metadata is retrieved when file upload has been registered
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
    When the file metadata is requested for the file "images/meme.jpg"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "images/meme.jpg",
      "is_publishable": true,
      "collection_id": "1234-asdfg-54321-qwerty",
      "title": "The latest Meme",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    """

  Scenario: The one where the user is not authorised to (pre)view
    Given I am not an authorised user
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
    When the file metadata is requested for the file "images/meme.jpg"
    Then the HTTP status code should be "403"


  Scenario: The one where the collection ID is not set
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When the file metadata is requested for the file "images/meme.jpg"
    Then I should receive the following JSON response with status "200":
    """
    {
      "path": "images/meme.jpg",
      "is_publishable": true,
      "title": "The latest Meme",
      "size_in_bytes": 14794,
      "type": "image/jpeg",
      "licence": "OGL v3",
      "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
      "state": "CREATED",
      "etag": ""
    }
    """

  Scenario: The file metadata is not found when file has not been registered
    Given I am an authorised user
    And the file "images/not-found.jpg" has not been registered
    When the file metadata is requested for the file "images/not-found.jpg"
    Then the HTTP status code should be "404"
