import jwt
import time

# make a JWT token usable in an "Authorization: Bearer XXX"
# header for talking to a local standalone search-service
def make_jwt():
    USER = "system.admin@opgtest.com"

    # this has to tally with the local/jwt-key secret set in
    # scripts/localstack/init/localstack_init.sh
    SECRET = "MyTestSecret"

    jwt_payload = {
        "session-data": USER,
        "iat": int(time.time())
    }

    return jwt.encode(jwt_payload, SECRET, algorithm="HS256")

if __name__ == "__main__":
    # output a header=value string suitable for use with
    # Pact's --header CLI arg
    print(f"Authorization=Bearer {make_jwt()}")
