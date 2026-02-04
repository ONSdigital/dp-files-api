Feature: Updating the content item in a files metadata

  Scenario: Updating the content item in the metadata of a registered file
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I update the content item of the file "images/meme.jpg" with:
      """
        {
          "content_item": {
            "dataset_id": "meme-dataset-2",
            "edition": "jan2",
            "version": "2"
          }
        }
      """
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
          "etag": "",
          "content_item": {
            "dataset_id": "meme-dataset-2",
            "edition": "jan2",
            "version": "2"
          }
        }
      """

  Scenario: Updating the content item in the metadata of a non registered file fails
    Given I am an authorised user
    When I update the content item of the file "images/non-existent-meme.jpg" with:
      """
        {
          "content_item": {
            "dataset_id": "meme-dataset-2",
            "edition": "jan2",
            "version": "2"
          }
        }
      """
    Then the HTTP status code should be "404"

  Scenario: An unauthorised user cannot update the content item of a registered file
    Given I am not an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I update the content item of the file "images/meme.jpg" with:
      """
        {
          "content_item": {
            "dataset_id": "meme-dataset-2",
            "edition": "jan2",
            "version": "2"
          }
        }
      """
    Then the HTTP status code should be "403"

  Scenario: Updating the content item in the metadata with an invalid content_item object fails
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I update the content item of the file "images/meme.jpg" with:
      """
        { invalid json }
      """
    Then the HTTP status code should be "400"

  Scenario: Only the content item in the metadata of a registered file is updated
    Given I am an authorised user
    And the file upload "images/meme.jpg" has been registered with:
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-21T15:13:14Z                                                      |
      | LastModified  | 2021-10-21T15:13:14Z                                                      |
      | State         | CREATED                                                                   |
    When I update the content item of the file "images/meme.jpg" with:
      """
        {
          "title": "An updated name",
          "state": "PUBLISHED",
          "content_item": {
            "dataset_id": "meme-dataset-3",
            "edition": "jan3",
            "version": "3"
          }
        }
      """
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
          "etag": "",
          "content_item": {
            "dataset_id": "meme-dataset-3",
            "edition": "jan3",
            "version": "3"
          }
        }
      """