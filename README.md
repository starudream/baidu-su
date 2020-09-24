# Baidu-Su

![Go](https://img.shields.io/github/workflow/status/starudream/baidu-su/Go/master?style=for-the-badge)
![Docker](https://img.shields.io/github/workflow/status/starudream/baidu-su/Docker/master?style=for-the-badge)
![License](https://img.shields.io/badge/License-Apache%20License%202.0-blue?style=for-the-badge)

## Config

```json
{
  "access_key": "xxx",
  "secret_key": "xxx",
  "bduss": "xxx",
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

![Version](https://img.shields.io/docker/v/starudream/baidu-su?style=for-the-badge)
![Size](https://img.shields.io/docker/image-size/starudream/baidu-su/latest?style=for-the-badge)
![Pull](https://img.shields.io/docker/pulls/starudream/baidu-su?style=for-the-badge)

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
