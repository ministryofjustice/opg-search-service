# Client for search-service
import logging
import requests
import time

from jwt_maker import make_jwt

class Client:

    PERSONS_INDEX_PATH = "/persons"
    PERSONS_SEARCH_PATH = "/persons/search"
    RETRY_WAIT_SECONDS = 1

    def __init__(self, search_service_url, logger):
        self.search_service_url = search_service_url
        self.logger = logger
        self.jwt = make_jwt()

    def _post(self, path, data):
        url = f"{self.search_service_url.rstrip('/')}/{path.lstrip('/')}"

        self.logger.warning(f"Making request to search service at URL {url} with data:")
        self.logger.warning(data)

        headers = {
            "Authorization": f"Bearer {self.jwt}",
            "Content-Type": "application/json"
        }

        return requests.post(url, headers=headers, json=data)

    def index_persons(self, data):
        return self._post(self.PERSONS_INDEX_PATH, data)

    # search for a person by term, retrying up to <retries> times, sleeping inbetween
    def search_persons(self, term, retries=0):
        data = {
            "term": term,
            "size": 10,
            "from": 0
        }

        response = self._post(self.PERSONS_SEARCH_PATH, data)

        if retries > 0 and (response.status_code != 200 or len(response.json()["results"]) == 0):
            time.sleep(self.RETRY_WAIT_SECONDS)
            return self.search_persons(term, retries - 1)

        return response
