
simple daemon to execute preconfigured commands over http


Example configuration, instructd.yaml

    http_addr: 0.0.0.0:8945
    auth:
      foobar:
        secret_key: asdf

    verbose: false
    commands:
    - id: hostname
        cmd:
        - "hostname"
        - "-f"

    - id: desktop.lock
        shell: |
        xdotool key "Control_L+Alt_L+l"


Run instructd

     instructd -config ./examples/instructd.yaml

Interact

     TOKEN="$(./examples/jwt foobar asdf )"

     curl -H "Authorization: bearer $TOKEN" -s -d '{"id": "hostname"}' http://localhost:8944/ | jq .

GET method is also supported (but not all params)

     curl "http://localhost:8945/?token=$TOKEN&id=hostname"
     
# JWT Authentication

a JWT is required to access the API. In this example, the access key is `foobar` and the secret key is `asdf` 

See examples/jwt for generating a JWT token in pure bash (It would be preferred to use a better library)

