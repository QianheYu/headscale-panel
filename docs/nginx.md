# This is a sample configuration of nginx

<p style="color: red">If you need https please turn it on in nginx, not in headscale and headscale-panel.</p>

This is a http example
```
server {
    listen 80;
    server_name example.com;
    
    # headscale-panel
    location ~ ^/(api|\.well-known/openid-configuration) {
        proxy_pass http://localhost:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    # headscale
    location ~ ^/(health|oidc|windows|apple|key|register|drep|bootstrap-dns|swagger|ts2021) {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```