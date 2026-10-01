#!/usr/bin/env bash
BASENAME=$(basename "$0")
BASEDIR=$(dirname "$0")

cd ${BASEDIR} || exit 1

EXT_FILE=${BASEDIR}/server-ext-ip.txt
CA_FILE=${BASEDIR}/root_cert.conf
SERVER=${BASEDIR}/server
CLIENT=${BASEDIR}/client
ROOT=${BASEDIR}/root
DB=${BASEDIR}/db
PRIVATE=${BASEDIR}/root/private
CERT=${BASEDIR}/root/certs

CN_NAME=CHINA
SERVER_CONF=server_cert.conf

# mkdir new dir
function createDir() {
if [ ! -d ${SERVER} ]; then
    mkdir -p ${SERVER}
fi

if [ ! -d ${CLIENT} ]; then
    mkdir -p ${CLIENT}
fi

if [ ! -d ${ROOT} ]; then
    mkdir -p ${ROOT}
fi
if [ ! -d ${DB} ]; then
    mkdir -p ${DB}
fi
touch ${DB}/{index,serial,crlnumber}
openssl rand -hex 16 > ${DB}/serial
if [ ! -d ${PRIVATE} ]; then
    mkdir -p ${PRIVATE}
fi
if [ ! -d ${CERT} ]; then
    mkdir -p ${CERT}
fi
}


function genConfigFile(){
cat << EOF > ${SERVER}/${SERVER_CONF}
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = IP:192.168.20.21,IP:127.0.0.1,IP:192.168.20.20, IP:192.168.20.22, IP:fd00::4, DNS:*.lson.se
EOF
}

function genPrivatekey(){
  openssl genrsa -des3 -out ${ROOT}/root.key 2048
  openssl genrsa -des3 -out ${SERVER}/server.key 2048
  openssl genrsa -des3 -out ${CLIENT}/client.key 2048
}


function genCsr(){
  openssl req -new -key ${ROOT}/root.key -out ${ROOT}/root.csr -subj "/C=CN/ST=GZ/L=GZ/O=lson/OU=lson/CN=${CN_NAME}"
  openssl req -new -key ${CLIENT}/client.key -out ${CLIENT}/client.csr -subj "/C=CN/ST=GZ/L=GZ/O=lson/OU=lson/CN=${CN_NAME}"
  openssl req -new -key ${SERVER}/server.key -out ${SERVER}/server.csr -subj "/C=CN/ST=GZ/L=GZ/O=lson/OU=lson/CN=${CN_NAME}"
}

function genPrivateAndCsr(){
    openssl req -out ${ROOT}/root.csr -config ${CA_FILE} -new  -newkey rsa:2048 -nodes -keyout ${PRIVATE}/root.key -subj "/C=CN/ST=GZ/L=GZ/O=lson/OU=lson/CN=${CN_NAME}"
    openssl req -out ${CLIENT}/client.csr  -new  -newkey rsa:2048 -nodes -keyout ${CLIENT}/client.key -subj "/C=CN/ST=GZ/L=GZ/O=lson/OU=lson/CN=${CN_NAME}"
    openssl req -out ${SERVER}/server.csr  -new  -newkey rsa:2048 -nodes -keyout ${SERVER}/server.key -subj "/C=CN/ST=GZ/L=GZ/O=lson/OU=lson/CN=${CN_NAME}"
}

function genCASign(){
    openssl  ca -batch  -notext -selfsign  -in ${ROOT}/root.csr -days 1000 -config ${CA_FILE} -extensions ca_ext -keyfile ${PRIVATE}/root.key  -out ${ROOT}/root.crt
    openssl  x509  -req -in ${CLIENT}/client.csr -days 1000 -sha1 -CA ${ROOT}/root.crt -CAkey ${PRIVATE}/root.key -CAserial ${CLIENT}/client.srl  -CAcreateserial  -out ${CLIENT}/client.crt
    #openssl  x509  -req -in ${SERVER}/server.csr -days 1000  -CA ${ROOT}/root.crt -CAkey ${ROOT}/root.key -CAserial ${SERVER}/server.srl  -CAcreateserial -out ${SERVER}/server.crt -extfile ${EXT_FILE}

    openssl  x509  -req -in ${SERVER}/server.csr -days 1000  -CA ${ROOT}/root.crt -CAkey ${PRIVATE}/root.key -CAserial ${SERVER}/server.srl  -CAcreateserial -out ${SERVER}/server.crt -extfile ${SERVER}/${SERVER_CONF}
    #openssl  x509  -req -in ${SERVER}/server.csr -days 1000  -CA ${ROOT}/root.crt -CAkey ${ROOT}/root.key -CAserial ${SERVER}/server.srl  -CAcreateserial -out ${SERVER}/server.crt
}

function genPKCS(){
    openssl pkcs12 -export -in ${CLIENT}/client.crt -inkey ${CLIENT}/client.key -out ${CLIENT}/client.p12 -name sslclient -password pass:'123456'
    openssl pkcs12 -export -in ${SERVER}/server.crt -inkey ${SERVER}/server.key -out ${SERVER}/server.p12 -name sslserver -password pass:'123456'
}


function genKeyStore(){
    keytool -importkeystore -srckeystore ${CLIENT}/client.p12 -destkeystore ${CLIENT}/client.keystore -srcstoretype PKCS12 -srcstorepass '123456' -deststoretype JKS -deststorepass '123456' -srcalias sslclient -destalias sslclient
    keytool -importkeystore -srckeystore ${SERVER}/server.p12 -destkeystore ${SERVER}/server.keystore -srcstoretype PKCS12 -srcstorepass '123456' -deststoretype JKS -deststorepass '123456' -srcalias sslserver -destalias sslserver
}

function genTrust(){
    keytool -import -trustcacerts -alias sslroot -file ${ROOT}/root.crt -keystore ${CLIENT}/clientTrust  -deststorepass '123456' -noprompt
    keytool -importcert  -alias sslserver -file ${SERVER}/server.crt -keystore ${CLIENT}/clientTrust -storepass '123456' -deststorepass '123456' -noprompt
    keytool -import -trustcacerts -alias sslroot -file ${ROOT}/root.crt -keystore ${SERVER}/serverTrust  -deststorepass '123456'   -noprompt
    keytool -importcert  -alias sslclient -file ${CLIENT}/client.crt -keystore ${SERVER}/serverTrust -storepass '123456' -deststorepass '123456'  -noprompt
}

function removeTmpFile(){
  rm  -f  ${ROOT}/*.conf
  rm  -f  ${ROOT}/*.csr
  rm  -f  ${SERVER}/*.conf
  rm  -f  ${SERVER}/*.csr
  rm  -f  ${SERVER}/*.srl
  rm  -f  ${SERVER}/*.key
  rm  -f  ${SERVER}/*.crt

  rm  -f  ${CLIENT}/*.conf
  rm  -f  ${CLIENT}/*.srl
  rm  -f  ${CLIENT}/*.csr
  rm  -f  ${CLIENT}/*.key
  rm  -f  ${CLIENT}/*.crt

  rm -rf ${ROOT}
  rm -rf ${SERVER}
  rm -rf ${CLIENT}
  rm -rf ${DB}
}


function usage() {
    echo "Usage: $BASENAME [options]"
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -t, --type      (pem,pkcs12,jks,all)Specify the type of certificate generation (default: all)"
    echo "  -c, --clean     Clean up generated files"
    exit 0
}


function main(){

    if [[ $# -eq 0 ]]; then
        echo "No arguments provided. Use -h or --help for usage information."
        exit 1
    fi

    createDir
    genConfigFile

    while [[ $# -gt 0 ]];  do
        case $1 in
            -h|--help)
                usage
                exit
                ;;
            -t|--type)
                TYPE="$2"
                case $TYPE in
                    pem)
                        echo "Generating certificates of type: $TYPE"
                        genPrivateAndCsr
                        genCASign
                        ;;
                    pkcs12)
                        echo "Generating certificates of type: $TYPE"
                        genPrivateAndCsr
                        genCASign
                        genPKCS
                        ;;
                    jks)
                        echo "Generating certificates of type: $TYPE"
                            genPrivateAndCsr
                            genCASign
                            genPKCS
                            genKeyStore
                            genTrust
                        ;;
                    all)
                        echo "Generating all certificate types"
                            genPrivateAndCsr
                            genCASign
                            genPKCS
                            genKeyStore
                            genTrust
                        ;;
                    *)
                        echo "Invalid type specified: $TYPE"
                        usage
                        ;;
                esac
                shift 2
                ;;
            -c|--clean)
                echo "Cleaning up generated files"
                removeTmpFile
                shift
                ;;
            *)
                echo "Unknown option: $1"
                usage
                ;;
        esac
    done

    #genPrivatekey
    #genCsr
    #removeTmpFile
}

main $@
