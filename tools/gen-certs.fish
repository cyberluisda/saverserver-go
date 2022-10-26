#!/usr/bin/fish

set -q ROOT_KEY_PASSPHRASE || set ROOT_KEY_PASSPHRASE "root1234"
set -q ROOT_CERT_DAYS || set ROOT_CERT_DAYS "36500"
set -q ROOT_COUNTRY || set ROOT_COUNTRY "SP"
set -q ROOT_STATE || set ROOT_STATE "Madrid"
set -q ROOT_ORG || set ROOT_ORG "Local"
set -q ROOT_COMMON_NAME || set ROOT_COMMON_NAME "Root"
set -q ROOT_EMAIL_ADDR || set ROOT_EMAIL_ADDR "root@local.local"

set -q INTCA_KEY_PASSPHRASE || set INTCA_KEY_PASSPHRASE "ca1234"
set -q INTCA_CERT_DAYS || set INTCA_CERT_DAYS "36500"
set -q INTCA_COUNTRY || set INTCA_COUNTRY "SP"
set -q INTCA_STATE || set INTCA_STATE "Madrid"
set -q INTCA_ORG || set INTCA_ORG "Local"
set -q INTCA_COMMON_NAME || set INTCA_COMMON_NAME "ca"
set -q INTCA_EMAIL_ADDR || set INTCA_EMAIL_ADDR "ca@local.local"

set -q SERVER_FQDN || set SERVER_FQDN "192.168.56.1"
set -q SERVER_CERT_DAYS || set SERVER_CERT_DAYS "36500"
set -q SERVER_KEY_SIZE || set SERVER_KEY_SIZE 2048
set -q SERVER_KEY_PASSPHRASE || set SERVER_KEY_PASSPHRASE "server1234"
set -q SERVER_COUNTRY || set SERVER_COUNTRY "SP"
set -q SERVER_STATE || set SERVER_STATE "Madrid"
set -q SERVER_ORG || set SERVER_ORG "Local"
set -q SERVER_COMMON_NAME || set SERVER_COMMON_NAME $SERVER_FQDN
set -q SERVER_ALT_NAME_PREFIX || set SERVER_ALT_NAME_PREFIX "IP"
set -q SERVER_EMAIL_ADDR || set SERVER_EMAIL_ADDR "server@local.local"

set -q USER_NAME || set USER_NAME "LuisDa"
set -q USER_CERT_DAYS || set USER_CERT_DAYS "36500"
set -q USER_KEY_SIZE || set USER_KEY_SIZE 2048
set -q USER_KEY_PASSPHRASE || set USER_KEY_PASSPHRASE "user1234"
set -q USER_COUNTRY || set USER_COUNTRY "SP"
set -q USER_STATE || set USER_STATE "Madrid"
set -q USER_ORG || set USER_ORG "Local"
set -q USER_COMMON_NAME || set USER_COMMON_NAME $USER_NAME
set -q USER_EMAIL_ADDR || set USER_EMAIL_ADDR "$USER_NAME@local.local"

set -q USER_KEY_REM_PASSPHRASE || set USER_KEY_REM_PASSPHRASE "yes"
set -q SERVER_KEY_REM_PASSPHRASE || set SERVER_KEY_REM_PASSPHRASE "yes"

function mkPaths
    mkdir -p tls/certs tls/private tls/crl tls/newcerts tls/intermediate
    mkdir -p \
        tls/intermediate/certs \
        tls/intermediate/crl \
        tls/intermediate/csr \
        tls/intermediate/newcerts \
        tls/intermediate/private

        echo 01 > tls/serial
    echo 1000 > tls/intermediate/serial

    touch tls/index.txt
    touch tls/intermediate/index.txt

    echo 1000 > tls/intermediate/crlnumber
end


function mkSslConfFile
    # From https://jamielinux.com/docs/openssl-certificate-authority/create-the-root-pair.html

    echo "(++) Creating tls/openssl.cnf"
    set mePath (pwd)
    echo '
[ ca ]
# `man ca`
default_ca = CA_default

[ CA_default ]
# Directory and file locations.
dir               = '$mePath'/tls
certs             = $dir/certs
crl_dir           = $dir/crl
new_certs_dir     = $dir/newcerts
database          = $dir/index.txt
serial            = $dir/serial
RANDFILE          = $dir/private/.rand
database          = $dir/index.txt

# The root key and root certificate.
private_key       = $dir/private/ca.key.pem
certificate       = $dir/certs/ca.cert.pem

# For certificate revocation lists.
crlnumber         = $dir/crlnumber
crl               = $dir/crl/ca.crl.pem
crl_extensions    = crl_ext
default_crl_days  = 30

# SHA-1 is deprecated, so use SHA-2 instead.
default_md        = sha256

name_opt          = ca_default
cert_opt          = ca_default
default_days      = 36500
preserve          = no
policy            = policy_strict

[ policy_strict ]
# The root CA should only sign intermediate certificates that match.
# See the POLICY FORMAT section of `man ca`.
countryName             = match
stateOrProvinceName     = match
organizationName        = match
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional

[ policy_loose ]
# Allow the intermediate CA to sign a more diverse range of certificates.
# See the POLICY FORMAT section of the `ca` man page.
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional

[ req ]
# Options for the `req` tool (`man req`).
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only

# SHA-1 is deprecated, so use SHA-2 instead.
default_md          = sha256

# Extension to add when the -x509 option is used.
x509_extensions     = v3_ca

[ req_distinguished_name ]
# See <https://en.wikipedia.org/wiki/Certificate_signing_request>.
countryName                     = Country Name (2 letter code)
stateOrProvinceName             = State or Province Name
0.organizationName              = Organization Name
commonName                      = Common Name
emailAddress                    = Email Address

# Optionally, specify some defaults.
countryName_default             = GB
stateOrProvinceName_default     = England
localityName_default            =
0.organizationName_default      = Alice Ltd
organizationalUnitName_default  =
emailAddress_default            = user@foo.com

[ v3_ca ]
# Extensions for a typical CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign

[ v3_intermediate_ca ]
# Extensions for a typical intermediate CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign

[ usr_cert ]
# Extensions for client certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = client, email
nsComment = "OpenSSL Generated Client Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, emailProtection

[ server_cert ]
# Extensions for server certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = server
nsComment = "OpenSSL Generated Server Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer:always
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[ crl_ext ]
# Extension for CRLs (`man x509v3_config`).
authorityKeyIdentifier=keyid:always

[ ocsp ]
# Extension for OCSP signing certificates (`man ocsp`).
basicConstraints = CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, digitalSignature
extendedKeyUsage = critical, OCSPSigning' \
        > tls/openssl.cnf

    echo "(++) Creating tls/intermediate/openssl.cnf"
    cp tls/openssl.cnf tls/intermediate
    sed -i -E 's|^(dir .*)$|\1/intermediate\npolicy            = policy_loose|' tls/intermediate/openssl.cnf
    sed -i -E 's|^(private_key       = .*)/ca\.key\.pem|\1/intermediate.key.pem|' tls/intermediate/openssl.cnf
    sed -i -E 's|^(certificate       = .*)/ca\.cert\.pem|\1/intermediate.cert.pem|' tls/intermediate/openssl.cnf
    sed -i '21,35d' tls/intermediate/openssl.cnf
    sed -i -E 's|^(. server_cert.*)$|\1\nsubjectAltName = '$SERVER_ALT_NAME_PREFIX':'$SERVER_FQDN'\n|' tls/intermediate/openssl.cnf
end

function rootCAKeyAndCert
    echo "(++) Creating tls/private/ca.key.pem"
    openssl genrsa -aes256 -passout pass:"$ROOT_KEY_PASSPHRASE" -out tls/private/ca.key.pem 4096
    chmod 400 tls/private/ca.key.pem

    echo "(++) Creating tls/certs/ca.cert.pem"
    openssl req -config tls/openssl.cnf \
        -passin pass:"$ROOT_KEY_PASSPHRASE" \
        -key tls/private/ca.key.pem \
        -new -x509 -days $ROOT_CERT_DAYS -sha256 -extensions v3_ca \
        -subj "/C=$ROOT_COUNTRY/ST=$ROOT_STATE/O=$ROOT_ORG/CN=$ROOT_COMMON_NAME/emailAddress=$ROOT_EMAIL_ADDR/" \
        -out tls/certs/ca.cert.pem

    chmod 444 tls/certs/ca.cert.pem

    echo "(++) Info tls/certs/ca.cert.pem"
    openssl x509 -noout -text -in tls/certs/ca.cert.pem
end

function intCAKeyCSRAndCert
    echo "(++) Creating tls/intermediate/private/intermediate.key.pem"
    openssl genrsa -aes256 -passout pass:"$INTCA_KEY_PASSPHRASE" -out tls/intermediate/private/intermediate.key.pem 4096
    chmod 400 tls/intermediate/private/intermediate.key.pem

    echo "(++) Creting tls/intermediate/csr/intermediate.csr.pem"
    openssl req -config tls/intermediate/openssl.cnf -new -sha256 \
        -passin pass:"$INTCA_KEY_PASSPHRASE" \
        -key tls/intermediate/private/intermediate.key.pem \
        -subj "/C=$INTCA_COUNTRY/ST=$INTCA_STATE/O=$INTCA_ORG/CN=$INTCA_COMMON_NAME/emailAddress=$INTCA_EMAIL_ADDR/" \
        -out tls/intermediate/csr/intermediate.csr.pem

    echo "(++) Creating tls/intermediate/certs/intermediate.cert.pem"
    openssl ca -config tls/openssl.cnf -extensions v3_intermediate_ca -batch \
        -passin pass:"$ROOT_KEY_PASSPHRASE" \
        -days $INTCA_CERT_DAYS -notext -md sha256 \
        -in tls/intermediate/csr/intermediate.csr.pem \
        -out tls/intermediate/certs/intermediate.cert.pem

    chmod 444 tls/intermediate/certs/intermediate.cert.pem

    echo "(++) Info tls/intermediate/certs/intermediate.cert.pem"
    openssl x509 -noout -text -in tls/intermediate/certs/intermediate.cert.pem

    echo "(++) Verifying tls/intermediate/certs/intermediate.cert.pem"
    openssl verify -CAfile tls/certs/ca.cert.pem tls/intermediate/certs/intermediate.cert.pem

end

function certChainFile
    echo "(++) Creating tls/intermediate/certs/ca-chain.cert.pem"
    cat tls/intermediate/certs/intermediate.cert.pem \
        tls/certs/ca.cert.pem > tls/intermediate/certs/ca-chain.cert.pem
    chmod 444 tls/intermediate/certs/ca-chain.cert.pem
end

function serverKeyCert
    echo "(++) Creating tls/intermediate/private/$SERVER_FQDN.key.pem"
    openssl genrsa -aes256 -passout pass:"$SERVER_KEY_PASSPHRASE" \
        -out tls/intermediate/private/$SERVER_FQDN.key.pem $SERVER_KEY_SIZE
    chmod 400 tls/intermediate/private/$SERVER_FQDN.key.pem

    echo "(++) Creating tls/intermediate/csr/$SERVER_FQDN.csr.pem CN=$SERVER_COMMON_NAME"
    openssl req -config tls/intermediate/openssl.cnf \
        -passin pass:"$SERVER_KEY_PASSPHRASE" \
        -key tls/intermediate/private/$SERVER_FQDN.key.pem \
        -subj "/C=$SERVER_COUNTRY/ST=$SERVER_STATE/O=$SERVER_ORG/CN=$SERVER_COMMON_NAME/emailAddress=$SERVER_EMAIL_ADDR/" \
        -new -sha256 -out tls/intermediate/csr/$SERVER_FQDN.csr.pem

    echo "(++) Creating tls/intermediate/certs/$SERVER_FQDN.cert.pem siged by intermediate CA"
    openssl ca -config tls/intermediate/openssl.cnf -batch \
        -passin pass:"$INTCA_KEY_PASSPHRASE" \
        -extensions server_cert -days $SERVER_CERT_DAYS -notext -md sha256 \
        -in tls/intermediate/csr/$SERVER_FQDN.csr.pem \
        -out tls/intermediate/certs/$SERVER_FQDN.cert.pem
    chmod 444 tls/intermediate/certs/$SERVER_FQDN.cert.pem

    echo "(++) Info tls/intermediate/certs/$SERVER_FQDN.cert.pem"
    openssl x509 -noout -text -in tls/intermediate/certs/$SERVER_FQDN.cert.pem

    echo "(++) Verifying tls/intermediate/certs/$SERVER_FQDN.cert.pem"
    openssl verify -CAfile tls/intermediate/certs/ca-chain.cert.pem \
        tls/intermediate/certs/$SERVER_FQDN.cert.pem
end

function userKeyCert
    echo "(++) Creating tls/intermediate/private/$USER_NAME.key.pem"
    openssl genrsa -aes256 -passout pass:"$USER_KEY_PASSPHRASE" \
        -out tls/intermediate/private/$USER_NAME.key.pem $USER_KEY_SIZE
    chmod 400 tls/intermediate/private/$USER_NAME.key.pem

    echo "(++) Creating tls/intermediate/csr/$USER_NAME.csr.pem CN=$USER_COMMON_NAME"
    openssl req -config tls/intermediate/openssl.cnf \
        -passin pass:"$USER_KEY_PASSPHRASE" \
        -key tls/intermediate/private/$USER_NAME.key.pem \
        -subj "/C=$USER_COUNTRY/ST=$USER_STATE/O=$USER_ORG/CN=$USER_COMMON_NAME/emailAddress=$USER_EMAIL_ADDR/" \
        -new -sha256 -out tls/intermediate/csr/$USER_NAME.csr.pem


    echo "(++) Creating tls/intermediate/certs/$USER_NAME.cert.pem siged by intermediate CA"
    openssl ca -config tls/intermediate/openssl.cnf -batch \
        -passin pass:"$INTCA_KEY_PASSPHRASE" \
        -extensions usr_cert -days $USER_CERT_DAYS -notext -md sha256 \
        -in tls/intermediate/csr/$USER_NAME.csr.pem \
        -out tls/intermediate/certs/$USER_NAME.cert.pem
    chmod 444 tls/intermediate/certs/$USER_NAME.cert.pem

    echo "(++) Info tls/intermediate/certs/$USER_NAME.cert.pem"
    openssl x509 -noout -text -in tls/intermediate/certs/$USER_NAME.cert.pem

    echo "(++) Verifying tls/intermediate/certs/$USER_NAME.cert.pem"
    openssl verify -CAfile tls/intermediate/certs/ca-chain.cert.pem \
        tls/intermediate/certs/$USER_NAME.cert.pem
end

function removePassPhrases
    if test "$USER_KEY_REM_PASSPHRASE" = "yes"
        echo "(++) Removing passphrase from tls/intermediate/private/$USER_NAME.key.pem"
        mv -f tls/intermediate/private/$USER_NAME.key.pem tls/intermediate/private/$USER_NAME.key-orig.pem
        openssl rsa \
            -passin pass:"$USER_KEY_PASSPHRASE"\
            -in tls/intermediate/private/$USER_NAME.key-orig.pem \
            -out tls/intermediate/private/$USER_NAME.key.pem

        rm -f tls/intermediate/private/$USER_NAME.key-orig.pem
    else
        echo "(II) USER_KEY_REM_PASSPHRASE is not yes. User key maintained with passphrase"
    end

    if test "$SERVER_KEY_REM_PASSPHRASE" = "yes"
        echo "(++) Removing passphrase from tls/intermediate/private/$SERVER_FQDN.key.pem"
        mv -f tls/intermediate/private/$SERVER_FQDN.key.pem tls/intermediate/private/$SERVER_FQDN.key-orig.pem
        openssl rsa \
            -passin pass:"$SERVER_KEY_PASSPHRASE"\
            -in tls/intermediate/private/$SERVER_FQDN.key-orig.pem \
            -out tls/intermediate/private/$SERVER_FQDN.key.pem

        rm -f tls/intermediate/private/$SERVER_FQDN.key-orig.pem
    else
        echo "(II) SERVER_KEY_REM_PASSPHRASE is not yes. Server key maintained with passphrase"
    end
end

function helpConfigCerts
    echo 'This is a resume of the keys and certificates generated

    tls/intermediate/private/'$SERVER_FQDN'.key.pem: The key to configure in the server. For example in -tlskey option of ../examples/netcat.go
    tls/intermediate/certs/'$SERVER_FQDN'.cert.pem: The cert to configure in the server, For example in -tlscert option of ../examples/netcat.go

    tls/intermediate/certs/ca-chain.cert.pem: The chain-CA file to configure in the client.
        For example create a new x509.CertPool, with x509.NewCertPool(), and set it to RootCAs property of tls.Config in golang client.
    tls/intermediate/private/'$USER_NAME'.key.pem: The User key file to configure in the client.
        For example set in keyFile param of tls.LoadX509KeyPair(certFile, keyFile) and then add the certficiate returned to Certificates
        property of tls.Config in a golang client
    tls/intermediate/certs/'$USER_NAME'.cert.pem: The User certficate file to configure in the client.
        For example set in certFile param of tls.LoadX509KeyPair(certFile, keyFile) and then add the certficiate returned to Certificates
        property of tls.Config in a golang client

Examples of run with examples/netcat.go as server and openssl as client

examples/netcat. (Run in root project path)
    go run examples/netcat.go -display \
        -tlscert tools/tls/intermediate/certs/'$SERVER_FQDN'.cert.pem \
        -tlskey tools/tls/intermediate/private/'$SERVER_FQDN'.key.pem \
        -user --address tcp://0.0.0.0:13000

openssl client (Run in root project path)
    echo "now is "(date) | openssl s_client -connect '$SERVER_FQDN':13000 \
        -key tools/tls/intermediate/private/'$USER_NAME'.key.pem \
        -cert tools/tls/intermediate/certs/'$USER_NAME'.cert.pem \
        -verify 9 -CAfile tools/tls/intermediate/certs/ca-chain.cert.pem
'
end

function displayHelp
    echo '
Usage: '(status filename)' [ -h | --help ] [ -c | --displayconfig ] [ -o | --displayconfigonly ]

This fish-shell script tool generate:
    * a rootCA key and certificate
    * an intermediate CA key and certificate signed with root CA certificate
    * a key and certificate, signed with intermediate CA, to configure in tls server
    * a key and cettificate, signed with intermediate CA, to identify a tls client.
    * a chain CA certificate that includes root and intermediate CA, to client trusts in the certificate exposed by the server.

The tool to create the certificates is openssl. It must be isntalled in the system.

All the available configuration opotions to create the certifcates are loaded from environemnet variables. Run with -o option to display
the available configuration options.

The output files is saved under `tls` path created in current working directory.

OPTIONS:

* `-h` / `--help`: Display current text
* `-c` / `--displayconfig`: If present the config options with values will be displayed at the begining of the program
* `-o` / `--displayconfigonly`: If present the config will be displayed in similar mannert that setting -c but then exit.

'
end

function displayConfig
    echo 'Valid environment variables to change the congiruation with current values are:

ROOT_KEY_PASSPHRASE: Passphrase used to encrypt key to generate root CA certificate. Current value: "'$ROOT_KEY_PASSPHRASE'"
ROOT_CERT_DAYS: Expiration time in days of the root CA certificate. Current value: "'$ROOT_CERT_DAYS'"
ROOT_COUNTRY: Country code (2 chars) of the root CA certificate. Current value: "'$ROOT_COUNTRY'"
ROOT_STATE: State of the country specified in the root CA certificate. Current value: "'$ROOT_STATE'"
ROOT_ORG: Organization specified in the root CA certificate. Current value: "'$ROOT_ORG'"
ROOT_COMMON_NAME: Common name (CN) specified in the root CA certificate. Current value: "'$ROOT_COMMON_NAME'"
ROOT_EMAIL_ADDR: Email speciffied in the root CA certificate. Current value: "'$ROOT_EMAIL_ADDR'"
INTCA_KEY_PASSPHRASE: Passphrase used to encrypt key to generate intermediate CA certificate. Current value: "'$INTCA_KEY_PASSPHRASE'"
INTCA_CERT_DAYS: Expiration time in days of the intermediate CA certificate. Current value: "'$INTCA_CERT_DAYS'"
INTCA_COUNTRY: Country code (2 chars) of the intermediate CA certificate. Current value: "'$INTCA_COUNTRY'"
INTCA_STATE: State of the country specified in the intermediate CA certificate. Current value: "'$INTCA_STATE'"
INTCA_ORG: Organization specified in the intermediate CA certificate. Current value: "'$INTCA_ORG'"
INTCA_COMMON_NAME: Common name (CN) specified in the intermediate CA certificate. Current value: "'$INTCA_COMMON_NAME'"
INTCA_EMAIL_ADDR: Email speciffied in the intermediate CA certificate. Current value: "'$INTCA_EMAIL_ADDR'"
SERVER_FQDN: Full qualified hostname of IP of the server to set in its certificate .Current value: "'$SERVER_FQDN'"
SERVER_CERT_DAYS: Expiration time in days of the server certificate. Current value: "'$SERVER_CERT_DAYS'"
SERVER_KEY_SIZE: Key size generated and used to create server certificate. Current value: "'$SERVER_KEY_SIZE'"
SERVER_KEY_PASSPHRASE: Passphrase used to encrypt key to generate server certificate.
    The key can be unencrypted at the end of the process if value of SERVER_KEY_REM_PASSPHRASE is `yes`. Current value: "'$SERVER_KEY_PASSPHRASE'"
SERVER_COUNTRY: Country code (2 chars) of the server certificate. Current value: "'$SERVER_COUNTRY'"
SERVER_STATE: State of the country specified in the server certificate. Current value: "'$SERVER_STATE'"
SERVER_ORG: Organization specified in the server certificate. Current value: "'$SERVER_ORG'"
SERVER_COMMON_NAME: Common name (CN) specified in the server certificate. Current value: "'$SERVER_COMMON_NAME'"
SERVER_ALT_NAME_PREFIX: Prefix (IP or DNS) pre-pend to the subject alternative name (SAN) specified in the server certificate. Current value: "'$SERVER_ALT_NAME_PREFIX'"
SERVER_EMAIL_ADDR: Email speciffied in the server certificate. Current value: "'$SERVER_EMAIL_ADDR'"
USER_NAME: Name to set in its certificate and used to set the name of key/certificate files. Current value: "'$USER_NAME'"
USER_CERT_DAYS: Expiration time in days of the user certificate. Current value: "'$USER_CERT_DAYS'"
USER_KEY_SIZE: Key size generated and used to create user certificate. Current value: "'$USER_KEY_SIZE'"
USER_KEY_PASSPHRASE Passphrase used to encrypt key to generate server certificate.
    The key can be unencrypted at the end of the process if value of USER_KEY_REM_PASSPHRASE is `yes`. Current value: "'$USER_KEY_PASSPHRASE'"
USER_COUNTRY: Country code (2 chars) of the user certificate. Current value: "'$USER_COUNTRY'"
USER_STATE: State of the country specified in the user certificate. Current value: "'$USER_STATE'"
USER_ORG: Organization specified in the user certificate. Current value: "'$USER_ORG'"
USER_COMMON_NAME: Common name (CN) specified in the user certificate.Current value: "'$USER_COMMON_NAME'"
USER_EMAIL_ADDR: Email speciffied in the user certificate. Current value: "'$USER_EMAIL_ADDR'"
USER_KEY_REM_PASSPHRASE: If value is `yes`, the user key is unencrypted and saved clear at the *end* of the process. Current value: "'$USER_KEY_REM_PASSPHRASE'"
SERVER_KEY_REM_PASSPHRASE: If value is `yes`, the server key is unencrypted and saved clear at the *end* of the process. Current value: "'$SERVER_KEY_REM_PASSPHRASE'"

    '
end


argparse 'h/help' 'c/displayconfig' 'o/displayconfigonly' -- $argv
or exit 1

set -q _flag_help
and begin
    displayHelp
    exit 0
end

set -q _flag_displayconfigonly
and begin
    displayConfig
    exit 0
end


set -q _flag_displayconfig
and displayConfig

echo "(**) Paths needed"
mkPaths
echo "(**) SSL config file"
mkSslConfFile
echo "(**) Root CA Key and Cert"
rootCAKeyAndCert
echo "(**) Intermediate CA Key and Cert signed by Root CA"
intCAKeyCSRAndCert
echo "(**) Certificate chain file (Root + intermediate)"
certChainFile
echo "(**) Server key and certificate"
serverKeyCert
echo "(**) User key and certificate"
userKeyCert
echo "(**) Removing passphrase from server and user keys"
removePassPhrases
echo "(**) Help text to configure certs"
helpConfigCerts