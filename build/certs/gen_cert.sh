#!/bin/sh

openssl req -new -key /certs/cert.key -subj "/CN=$1" -sha256 | openssl x509 -req -days 3650 -CA /certs/ca.crt -CAkey /certs/ca.key -set_serial "$2" > /certs/nck.crt

#for browser
#openssl req -new -nodes -keyout mail.ru.key -out mail.ru.csr -config req.cnf
#openssl x509 -req -in mail.ru.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out mail.ru.crt -days 365 -extensions v3_req -extfile req.cnf
