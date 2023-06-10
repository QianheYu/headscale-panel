以下是使用 OpenSSL 生成自签名证书的步骤：

1. 生成一个 2048 位 RSA 密钥（看完发现，其实这一步可以省略）：
```
# 私钥 server.key
openssl genrsa -out server.key 2048

# 公钥 server_pub.key
openssl rsa -in server.key -pubout -out server_pub.key
```

2. 创建签名请求：
```
# 请求 server.csr
openssl req -new -key server.key -out server.csr
```

3. 创建自签名证书：
```
# 证书 server.crt 和密钥 server.key
openssl x509 -req -days 3650 -in server.csr -signkey server.key -out server.crt
```

Source: Conversation with Bing, 4/21/2023
(1) 使用 openssl 创建自签名证书 - 码厩. https://bing.com/search?q=%e4%bd%bf%e7%94%a8openssl%e7%94%9f%e6%88%90%e8%87%aa%e7%ad%be%e5%90%8d%e8%af%81%e4%b9%a6.
(2) 使用 OpenSSL 生成自签名证书 - IBM. https://www.ibm.com/docs/zh/SSMNED_v10/com.ibm.apic.cmc.doc/task_apionprem_gernerate_self_signed_openSSL.html.
(3) Windows下使用OpenSSL生成自签名证书 - CSDN博客. https://blog.csdn.net/xuanyushifeng/article/details/104511095.
(4) 使用OpenSSL生成自签名证书相关命令 - CSDN博客. https://blog.csdn.net/fengbingchun/article/details/107386847.
(5) 使用自定义根 CA 生成自签名证书 - Azure Application Gateway. https://learn.microsoft.com/zh-cn/azure/application-gateway/self-signed-certificates.
(6) 使用OpenSSL生成自签名SSL证书_nklinsirui的博客-CSDN博客. https://blog.csdn.net/nklinsirui/article/details/89432430.