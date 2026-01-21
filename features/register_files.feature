Feature: Register new file upload

  Scenario: Register that a collection upload has started
    Given I am an authorised user
    When the file upload is registered with payload:
        """
        {
          "path": "images/meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    Then the HTTP status code should be "201"
    And the following document entry should be created:
      | Path          | images/meme.jpg                                                          |
      | IsPublishable | true                                                                      |
      | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-19T09:30:30Z                                                      |
      | LastModified  | 2021-10-19T09:30:30Z                                                      |
      | State         | CREATED                                                                   |
  
  Scenario: Register that a bundle upload has started
    Given I am an authorised user
    When the file upload is registered with payload:
        """
        {
          "path": "images/meme.jpg",
          "is_publishable": true,
          "bundle_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    Then the HTTP status code should be "201"
    And the following document entry should be created:
      | Path          | images/meme.jpg                                                           |
      | IsPublishable | true                                                                      |
      | BundleID      | 1234-asdfg-54321-qwerty                                                   |
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-19T09:30:30Z                                                      |
      | LastModified  | 2021-10-19T09:30:30Z                                                      |
      | State         | CREATED                                                                   |

  Scenario: Attempting to register a file with a path that is already register
    Given I am an authorised user
    And the file upload "images/old-meme.jpg" has been registered
    When the file upload is registered with payload:
        """
        {
          "path": "images/old-meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    Then the HTTP status code should be "409"

  Scenario: Attempting to register a file with a path that is already register
    Given I am not an authorised user
    When the file upload is registered with payload:
        """
        {
          "path": "images/meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    Then the HTTP status code should be "403"
  
  Scenario: Attempting to register a file with both collection_id and bundle_id in the metadata
    Given I am an authorised user
    When the file upload is registered with payload:
        """
        {
          "path": "images/meme.jpg",
          "is_publishable": true,
          "collection_id": "1234-asdfg-54321-qwerty",
          "bundle_id": "123-bundle-id",
          "title": "The latest Meme",
          "size_in_bytes": 14794,
          "type": "image/jpeg",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        }
        """
    Then the HTTP status code should be "400"

  Scenario: Register file with dataset content_item metadata
    Given I am an authorised user
    When the file upload is registered with payload:
        """
        {
          "path": "datasets/cpih/2024/data.csv",
          "is_publishable": true,
          "title": "CPIH Dataset 2024",
          "size_in_bytes": 54321,
          "type": "text/csv",
          "licence": "OGL v3",
          "licence_url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/",
          "content_item": {
            "dataset_id": "cpih-dataset-001",
            "edition": "2024",
            "version": "1"
          }
        }
        """
    Then the HTTP status code should be "201"
    And the following document entry should be created:
      | Path          | datasets/cpih/2024/data.csv                                               |
      | IsPublishable | true                                                                      |
      | Title         | CPIH Dataset 2024                                                         |
      | SizeInBytes   | 54321                                                                     |
      | Type          | text/csv                                                                  |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | DatasetID     | cpih-dataset-001                                                          |
      | Edition       | 2024                                                                      |
      | Version       | 1                                                                         |
      | CreatedAt     | 2021-10-19T09:30:30Z                                                      |
      | LastModified  | 2021-10-19T09:30:30Z                                                      |
      | State         | CREATED                                                                   |
