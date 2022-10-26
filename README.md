# saverserver-go

`saverserver` is a golang library to start simple tcp, udp, or other golang socket supported server and give access to the payload sent by the client during or after the server is up and running

# examples

There are several running and simple applications that are defined as examples to use the library. These apps are saved `examples` directory.

Run each example without arguments (or `-h`) option to get online help.

# tools

The tools provided with the project are:

* `tools/gen-certs.fish` fish-shell script to create root CA, intermediate CA, server and user certificates using openssl tool. Run with `-h` option to get the help and usage information.

# TODO

- [x] TLS Listener
- [x] Add callback support for each payload received
- [x] Helper tool/scripts to create certificates (server and client) in order to be a valid test mockup for `devo-sender` from [devo-go](https://github.com/cyberluisda/devo-go)