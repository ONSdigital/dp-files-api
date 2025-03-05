Feature: Optional Collection ID
  Scenario: The one where the collection ID is not sent with the file meta data
    Given I am an authorised user
    When the file upload is registered with payload:
        """
        {
          "path": "images/meme.jpg",
          "is_publishable": true,
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
      | Title         | The latest Meme                                                           |
      | SizeInBytes   | 14794                                                                     |
      | Type          | image/jpeg                                                                |
      | Licence       | OGL v3                                                                    |
      | LicenceURL    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
      | CreatedAt     | 2021-10-19T09:30:30Z                                                      |
      | LastModified  | 2021-10-19T09:30:30Z                                                      |
      | State         | CREATED                                                                   |