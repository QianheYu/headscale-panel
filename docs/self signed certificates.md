The following are the steps for generating a self-signed certificate using OpenSSL

1. Generate a CA certificate and key
```
openssl req -new -x509 -days 365 -nodes -out ca.crt -keyout ca.key
```

2. Generate the private and public keys
```
# public key server.csr, private key server.key
openssl req -new -newkey rsa:2048 -nodes -keyout server.key -out server.csr -subj "/CN=localhost:50443" -addext "subjectAltName = DNS:localhost:50443"
```

3. Sign the public key with a CA certificate
```
# certificate server.crt
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -extfile <(printf "subjectAltName=DNS:localhost:50443")
Certificate request self-signature ok
subject=CN = localhost:50443
```

4. Verify the certificate
```
openssl verify -CAfile /home/ubuntu/ssl/ca.crt /home/ubuntu/ssl/server.crt
/home/ubuntu/ssl/server.crt: OK
```