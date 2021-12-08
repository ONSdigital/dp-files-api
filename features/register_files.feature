Feature: Register new file upload

    Scenario: Register that an upload has started

        When the file upload is registered with payload:
        """
        {
          "path": "/images/meme.jpg",
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
            | Path          | /images/meme.jpg                                                          |
            | IsPublishable | true                                                                      |
            | CollectionID  | 1234-asdfg-54321-qwerty                                                   |
            | Title         | The latest Meme                                                           |
            | SizeInBytes   | 14794                                                                     |
            | Type          | image/jpeg                                                                |
            | Licence       | OGL v3                                                                    |
            | LicenceUrl    | http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/ |
            | CreatedAt     | 2021-10-19T09:01+00:00                                                    |
            | LastModified  | 2021-10-19T09:01+00:00                                                    |
            | State         | CREATED                                                                   |
