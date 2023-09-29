import json
import os
from flask import Flask, request, Response

from client import Client

app = Flask(__name__)

search_service_url = os.environ.get("SEARCH_SERVICE_URL", "")

client = Client(search_service_url, app.logger)

with open(os.path.join(os.path.dirname(__file__), "./persons.json")) as persons_file:
    persons = json.loads(persons_file.read())

@app.post("/provider_state_change")
def provider_state_change():
    # The following states are defined in Sirius pact consumer test
    # ApplicationTest\Service\SearchServicePactTest:
    #   * Person called Peter with person type type1 has been indexed
    #   * Person with UID 7000-8813-9100 has been indexed
    #   * Person called Jarrett Wisniowski born 1928-04-15 with postcode L1 8WH has been indexed
    # Note that tests which check indexing don't require provider state to be set up beforehand
    #
    # Incoming state change requests look like this:
    #   * setup state change = {'action': 'setup', 'params': {}, 'state': '<state name, see above>'}
    #   * teardown state change = {'action': 'teardown', 'params': {}, 'state': '<state name, see above>'}
    # The pact verifier is configured (in docker-compose.yml) to send a tear down request
    # after verifying each pact.

    # Note that state is present but set to "" if not specified in the consumer pact.
    state = request.get_json().get("state")
    if state is None:
        return "No state set in state change request", 400

    person = persons.get(state, None)
    headers = {"Content-Type": "text/plain; charset=utf-8"}

    if person is None:
        return "Unrecognised or empty state: IGNORED", 200, headers

    response = client.index_persons({"persons": [person]})
    if response.status_code != 202:
        return "Unable to index person", 500, headers

    response = client.search_persons(person["firstname"], retries=3)
    if response.status_code != 200:
        return "Unable to confirm person was indexed in a timely fashion", 500, headers

    return "Person was indexed", 200, headers

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5175)
