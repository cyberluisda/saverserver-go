# saverserver-go

`saverserver` is a golang library to start simple tcp, udp, or other golang socket supported server and give access to the payload sent by the client during or after the server is up and running

# tools

The tools provided with the project are:

* `tools/gen-certs.fish` fish-shell script to create root CA, intermediate CA, server and user certificates using openssl tool. Run with `-h` option to get the help and usage information.

# TODO

- [x] TLS Listener
- [x] Add callback support for each payload received
- [ ] Helper tool/scripts to create certificates (server and client) in order to be a valid test mockup for `devo-sender` from [devo-go](https://github.com/cyberluisda/devo-go)