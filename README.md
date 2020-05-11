# Baidu-Su

![Go](https://github.com/starudream/baidu-su/workflows/Go/badge.svg)
![Docker](https://github.com/starudream/baidu-su/workflows/Docker/badge.svg)
![License](https://img.shields.io/badge/License-Apache%20License%202.0-blue)

## Config

```json
{
  "access_key": "xxx",
  "secret_key": "xxx",
  "cron": "* 4 * * *",
  "timezone": "Asia/Shanghai",
  "certs": [
    {
      "domain": "52xckl.cn",
      "name": "52xckl.cn",
      "crt_path": "/ssl/*.52xckl.cn.crt",
      "key_path": "/ssl/*.52xckl.cn.key"
    }
  ]
}
```

## Usage

![Version](https://img.shields.io/docker/v/starudream/baidu-su)
![Size](https://img.shields.io/docker/image-size/starudream/baidu-su/latest)
![Pull](https://img.shields.io/docker/pulls/starudream/baidu-su)

```bash
docker pull starudream/baidu-su
```

```bash
docker run -d \
    --name baidu-su \
    --restart always \
    -e DEBUG=true \
    -v /opt/docker/baidu-su/config.json:/config.json \
    -v /usr/local/openresty/nginx/conf/ssl:/ssl:ro \
    starudream/baidu-su:latest
```

## License

[Apache License 2.0](./LICENSE)
