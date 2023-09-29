from flask import Flask, request

app = Flask(__name__)

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
    app.logger.warning(request.get_json())
    return request.data

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5175)
