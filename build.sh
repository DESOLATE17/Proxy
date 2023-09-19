cd ./build/certs
chmod +x ./gen_ca.sh
./gen_ca.sh
cp ca.crt /usr/local/share/ca-certificates
update-ca-certificates
cd ../..
docker compose build
docker compose up
